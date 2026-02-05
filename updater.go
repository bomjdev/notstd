package notstd

import (
	"context"
	"sync"
	"time"
)

// Source представляет источник данных для обновления.
// T - тип данных, который возвращает источник (обычно map[K]V).
type Source[T any] interface {
	// Fetch получает данные из источника.
	// Контекст используется для отмены, timeout и передачи метаданных.
	Fetch(ctx context.Context) (T, error)
}

// Sink представляет приемник данных - куда записываются обновления.
type Sink[K comparable, V any] interface {
	// Apply применяет обновления к хранилищу.
	// data содержит данные для обновления (может быть полный набор или частичный).
	Apply(ctx context.Context, data map[K]V) error
}

// UpdateStrategy определяет стратегию применения обновлений к хранилищу.
type UpdateStrategy int

const (
	// StrategyReplace - полная замена данных в хранилище.
	// Удаляет все существующие данные, записывает только то, что пришло в обновлении.
	// Использование: когда источник всегда возвращает полный набор данных.
	StrategyReplace UpdateStrategy = iota

	// StrategyMerge - объединение данных.
	// Обновляет существующие записи, добавляет новые, НЕ удаляет отсутствующие.
	// Использование: когда источник может возвращать частичные обновления.
	StrategyMerge

	// StrategyUpsertOnly - только вставка и обновление.
	// Аналогично StrategyMerge, но явно подчеркивает что удаления не происходит.
	// Использование: когда важно накапливать данные и не терять старые записи.
	StrategyUpsertOnly

	// StrategyIncremental - инкрементальное обновление с явной обработкой удалений.
	// Требует чтобы источник возвращал информацию о том, какие ключи нужно удалить.
	// Использование: когда источник возвращает дельту (добавления, обновления, удаления).
	StrategyIncremental
)

// StoreSink - стандартная реализация Sink для Store.
type StoreSink[K comparable, V any] struct {
	store    *Store[K, V]
	strategy UpdateStrategy
}

// NewStoreSink создает новый StoreSink с указанной стратегией обновления.
func NewStoreSink[K comparable, V any](store *Store[K, V], strategy UpdateStrategy) *StoreSink[K, V] {
	return &StoreSink[K, V]{
		store:    store,
		strategy: strategy,
	}
}

// Apply применяет данные к Store согласно выбранной стратегии.
func (s *StoreSink[K, V]) Apply(ctx context.Context, data map[K]V) error {
	s.store.Lock()
	defer s.store.Unlock()

	switch s.strategy {
	case StrategyReplace:
		// Полная замена: очищаем Store и записываем новые данные
		s.store.m = make(map[K]V, len(data))
		for k, v := range data {
			s.store.m[k] = v
		}

	case StrategyMerge, StrategyUpsertOnly:
		// Merge/Upsert: обновляем существующие и добавляем новые
		for k, v := range data {
			s.store.m[k] = v
		}

	case StrategyIncremental:
		// Для инкрементальных обновлений используется та же логика что и Merge
		// Специальная обработка удалений должна быть на уровне Source
		// (например, Source может передавать zero-value для удаления)
		for k, v := range data {
			s.store.m[k] = v
		}
	}

	return nil
}

//// Middleware - функция-обертка для добавления дополнительной логики к Fetch.
//// Примеры: retry, logging, validation, caching.
//type Middleware[T any] func(next FetchFunc[T]) FetchFunc[T]

// FetchFunc - сигнатура функции получения данных.
type FetchFunc[T any] func(ctx context.Context) (T, error)

func (ff FetchFunc[T]) Fetch(ctx context.Context) (T, error) { return ff(ctx) }

// Updater - основная структура для фонового обновления данных.
type Updater[K comparable, V any] struct {
	source   Source[map[K]V]
	sink     Sink[K, V]
	interval time.Duration

	middlewares []Middleware[FetchFunc[map[K]V]]

	// Колбэки для мониторинга и реакции на события
	onSuccess func(data map[K]V)
	onError   func(error)

	// Управление жизненным циклом
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Метаданные (опционально для мониторинга)
	mu          sync.RWMutex
	lastSuccess time.Time
	lastError   error
}

// NewUpdater создает новый Updater.
func NewUpdater[K comparable, V any](
	source Source[map[K]V],
	sink Sink[K, V],
	interval time.Duration,
) *Updater[K, V] {
	return &Updater[K, V]{
		source:      source,
		sink:        sink,
		interval:    interval,
		middlewares: make([]Middleware[FetchFunc[map[K]V]], 0),
	}
}

// WithMiddleware добавляет middleware в цепочку обработки.
// Middleware выполняются в порядке добавления.
func (u *Updater[K, V]) WithMiddleware(mw ...Middleware[FetchFunc[map[K]V]]) *Updater[K, V] {
	u.middlewares = append(u.middlewares, mw...)
	return u
}

// WithSuccessHandler устанавливает обработчик успешных обновлений.
func (u *Updater[K, V]) WithSuccessHandler(handler func(data map[K]V)) *Updater[K, V] {
	u.onSuccess = handler
	return u
}

// WithErrorHandler устанавливает обработчик ошибок.
func (u *Updater[K, V]) WithErrorHandler(handler func(error)) *Updater[K, V] {
	u.onError = handler
	return u
}

// Start запускает фоновое обновление данных.
// Возвращает управление немедленно, обновления происходят в горутине.
// Первое обновление произойдет после истечения interval.
func (u *Updater[K, V]) Start(ctx context.Context) {
	u.ctx, u.cancel = context.WithCancel(ctx)
	u.wg.Add(1)

	go u.run()
}

// StartSync запускает фоновое обновление данных с синхронным первым обновлением.
// Дожидается первого успешного fetch перед запуском фоновой горутины.
// Возвращает ошибку если первое обновление не удалось.
// После успешного первого обновления работает как Start().
func (u *Updater[K, V]) StartSync(ctx context.Context) error {
	u.ctx, u.cancel = context.WithCancel(ctx)

	// Выполняем первое обновление синхронно
	if err := u.updateOnce(u.ctx); err != nil {
		u.cancel()
		return err
	}

	// После успешного обновления запускаем фоновую горутину
	u.wg.Add(1)
	go u.run()

	return nil
}

// Stop останавливает фоновое обновление и ждет завершения текущей итерации.
func (u *Updater[K, V]) Stop() {
	if u.cancel != nil {
		u.cancel()
	}
	u.wg.Wait()
}

// run - основной цикл обновления.
func (u *Updater[K, V]) run() {
	defer u.wg.Done()

	ticker := time.NewTicker(u.interval)
	defer ticker.Stop()

	for {
		select {
		case <-u.ctx.Done():
			return
		case <-ticker.C:
			u.update()
		}
	}
}

// update выполняет одну итерацию обновления (без возврата ошибки).
func (u *Updater[K, V]) update() {
	_ = u.updateOnce(u.ctx)
}

// updateOnce выполняет одну итерацию обновления и возвращает ошибку.
func (u *Updater[K, V]) updateOnce(ctx context.Context) error {
	// Строим цепочку middleware
	fetchFunc := u.buildFetchChain()

	// Выполняем fetch через middleware chain
	data, err := fetchFunc(ctx)
	if err != nil {
		u.setLastError(err)
		if u.onError != nil {
			u.onError(err)
		}
		return err
	}

	// Применяем данные к sink
	if err := u.sink.Apply(ctx, data); err != nil {
		u.setLastError(err)
		if u.onError != nil {
			u.onError(err)
		}
		return err
	}

	// Успешное обновление
	u.setLastSuccess()
	if u.onSuccess != nil {
		u.onSuccess(data)
	}

	return nil
}

// buildFetchChain строит цепочку middleware для fetch.
func (u *Updater[K, V]) buildFetchChain() FetchFunc[map[K]V] {
	// Базовая функция - вызов source.Fetch
	fetchFunc := u.source.Fetch

	// Оборачиваем в middleware в обратном порядке
	// (чтобы первый добавленный middleware был внешним)
	for i := len(u.middlewares) - 1; i >= 0; i-- {
		fetchFunc = u.middlewares[i](fetchFunc)
	}

	return fetchFunc
}

// LastSuccess возвращает время последнего успешного обновления.
func (u *Updater[K, V]) LastSuccess() time.Time {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.lastSuccess
}

// LastError возвращает последнюю ошибку обновления.
func (u *Updater[K, V]) LastError() error {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.lastError
}

func (u *Updater[K, V]) setLastSuccess() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.lastSuccess = time.Now()
	u.lastError = nil
}

func (u *Updater[K, V]) setLastError(err error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.lastError = err
}
