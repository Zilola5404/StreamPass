# StreamPass — Полный сценарий SaaS-продукта

> Основано на: `docs/02_TZ.md` (§1–22), `docs/33_DirectVsVpnBypass.md`, текущий Android-клиент и Admin UI.  
> Дата: 2026-08-06  
> Позиционирование: **ускоритель / интеллектуальный роутер**, не «полный VPN» и не «ускорение интернета».  
> **Детальная FS (источник для Dev/QA):** `docs/02.2_FunctionalSpecification.md`  
> **Чеклист реализации + BL-040…:** `docs/35_FS_ImplementationChecklist.md`

Документ описывает, **что должен делать полноценный SaaS StreamPass**: экраны, кнопки, данные, системные реакции и фоновые процессы. Где поведение уже есть в MVP — отмечено; где нужно для зрелого SaaS — отмечено как **SaaS target**.

---

## 0. Продукт в одном абзаце

Пользователь устанавливает StreamPass, регистрируется, оплачивает подписку и нажимает **одну кнопку**. Система сама выбирает relay, применяет правила DIRECT/RELAY, обходит банки/госуслуги, обновляет правила и конфиг, переключается при сбоях и собирает только техническую телеметрию. Российские сервисы работают «как без VPN»; зарубежные (YouTube, Google, GitHub, AI…) — через Hysteria2 relay.

---

## 1. Роли

| Роль | Цель |
|------|------|
| **Гость** | Понять продукт, зарегистрироваться / войти |
| **Пользователь без подписки** | Видеть Home, но Connect заблокирован → оплата |
| **Подписчик** | Connect / Disconnect, выбор региона, настройки, диагностика |
| **Поддержка** | Читать диагностический лог (экспорт с устройства), статус подписки |
| **Оператор / Admin** | Users, Relays, Rules, Config, Health, мониторинг |
| **Система** | Health monitor, fallback портов, rule/config poll, billing webhook, backup |

---

## 2. Карта навигации клиента (Android)

```
[Установка APK / OTA]
        │
        ▼
 Onboarding (Войти / Регистрация)
        │  успех
        ▼
 ┌──────────────── Home ────────────────┐
 │  Orb Connect/Disconnect              │
 │  Карточка Relay + Ping               │
 │  Баннеры: нет сети / нет подписки    │
 │  Auto Mode (переключатель)           │
 └──────┬──────────┬──────────┬─────────┘
        │          │          │
   Статистика   Серверы   Настройки ──► Исключения
        │          │          │     ──► Приложения без VPN
   (заглушка)  регионы/     │     ──► Диагностика
               relay        │
                            └── Premium ──► Подписка (оплата / отмена)
```

Нижняя навигация Home: **Главная | Статистика | Серверы | Настройки**.

---

## 3. Сквозной happy-path (новый пользователь)

| # | Действие пользователя | Что видит / происходит |
|---|----------------------|-------------------------|
| 1 | Скачивает APK (сайт / OTA URL / sideload) | Установка Android-приложения |
| 2 | Первый запуск | Splash ≤ 2 с → **Onboarding** (нет сессии) |
| 3 | «Зарегистрироваться» → email + пароль → «Создать аккаунт» | `POST /register` → JWT access+refresh → переход на **Home** |
| 4 | Home: баннер «Подписка неактивна» | Connect недоступен (gate по ТЗ §22) |
| 5 | Иконка Premium / баннер → **Подписка** → «Оплатить» | `POST /payments` → браузер ЮKassa → webhook → статус ACTIVE |
| 6 | Возврат в приложение | Подписка обновлена: «Действует до …» |
| 7 | Нажатие орба **Подключить** | Запрос VPN permission (первый раз) → PrepareRelay (UDP/TCP fallback) → TUN up → статус **Connected** |
| 8 | Пользование сетью | `.ru` / банки / госуслуги — DIRECT / bypass; YouTube и т.п. — RELAY |
| 9 | Нажатие орба снова | Disconnect → TUN down → «Авто-маршрут готов» |
| 10 | (опционально) Настройки → Автоподключение | При следующем старте с активной подпиской — auto-connect |

---

## 4. Экраны клиента — детально

### 4.1 Onboarding (вход / регистрация)

**Данные на экране**
- Бренд: StreamPass  
- Подзаголовок (ценность: стабильный доступ к зарубежным сервисам без ручной настройки)  
- Поля: Email, Пароль  
- Текст ошибки (если есть)  
- Кнопка: **Войти** *или* **Создать аккаунт**  
- Переключатель режима: «Нет аккаунта? Зарегистрироваться» / «Уже есть аккаунт? Войти»

**При нажатии**

| Элемент | Поведение |
|---------|-----------|
| Войти | `POST /login` → сохранить токены (secure storage) → Home |
| Создать аккаунт | `POST /register` → токены → Home |
| Неверный пароль | Остаётся на Onboarding, текст ошибки (без утечки «email exists» в деталях — SaaS: единое сообщение) |
| Backend недоступен | Дружелюбное «Сервис временно недоступен» |
| Сессия протухла на старте | Refresh → fail → Onboarding |

**SaaS target (дополнить):** «Забыли пароль?», политика пароля, Terms/Privacy ссылки, SSO опционально позже (не в MVP §21).

---

### 4.2 Home (главный экран) — ТЗ §20

**Данные**
| Блок | Содержимое |
|------|------------|
| Шапка | Меню (≡) → Настройки; заголовок StreamPass; иконка Premium |
| Статус-чип | «Система активна» (connected) / «Авто-маршрут готов» (idle) / ошибка |
| Орб | Состояние: disconnected / connecting / connected / error; при connected — длительность сессии |
| Relay-карточка | Флаг региона, название (Amsterdam NL…), id relay (`nl-native-1`), Ping N ms |
| Route-карточка | Текст про smart routing + переключатель **Auto Mode** |
| Баннеры | Нет связи с API («Повторить»); подписка неактивна (CTA оплатить) |
| Низ | Главная \| Статистика \| Серверы \| Настройки |

**При нажатии**

| Элемент | Условие | Поведение |
|---------|---------|-----------|
| Орб → Connect | Подписка ACTIVE + есть relay | `VpnService.prepare` → permission dialog → PrepareRelay (fallback UDP 443→8443→24443, TCP underlay) → StartTunnel → Connected + ping |
| Орб → Connect | Нет подписки | SnackBar + переход на **Подписка** |
| Орб → Connect | Нет relay / API ошибка | Состояние error + сообщение «Не удалось загрузить серверы» |
| Орб → Disconnect | Connected | Stop tunnel → Disconnected |
| Relay-карточка | — | Открыть **Серверы** |
| Premium / баннер подписки | — | **Подписка** |
| Auto Mode switch | — | Локальный UI-флаг режима (в зрелом SaaS должен синхронизироваться с «Автовыбор Relay» в настройках) |
| Повторить | API down | Повтор bootstrap: config, subscription, servers |
| Диалог обновления | `min_supported` / `latest` | «Позже» / «Скачать» (OTA URL) |

**Система без нажатий (после Connect)**
1. Decision Engine: Domain/CIDR/User exclusions → DIRECT \| RELAY \| FALLBACK  
2. Split-tunnel: RU IPv4 вне TUN  
3. App bypass: банки/госуслуги без `TRANSPORT_VPN`  
4. Split DNS: `.ru` → Yandex; foreign → DoH  
5. Rule poll + hot-reload без переустановки  
6. Relay health / degradation → переключение (ТЗ §2, §22)  
7. Telemetry: RTT, loss, relay id, version, OS, connect ms, error code — **без URL и истории**

---

### 4.3 Серверы / Регионы

**Данные**
- Pull-to-refresh  
- Плитка **Автовыбор** (отмечена, если включён)  
- Группы регионов: DE / NL / PL / FI (каталог + `region_name`)  
- Внутри: `relay id`, healthy / RTT ms или «Недоступен»

**При нажатии**

| Элемент | Поведение |
|---------|-----------|
| Автовыбор | `autoSelectRelay=true`, сброс preferred server → pop → Home перечитывает servers, pickBestRelay |
| Заголовок региона | Предпочтительный регион, авто внутри региона → Home |
| Конкретный relay | `autoSelectRelay=false`, preferredServerId → Home; если уже connected — **SaaS:** reconnect к новому relay |
| Pull refresh | `GET /servers` |

**SaaS target:** показать load/latency тренд; «рекомендуем» badge; запрет выбора unhealthy с понятным текстом.

---

### 4.4 Подписка (Billing — ТЗ §15)

**Данные**
- Статус: ACTIVE / неактивна  
- «Действует до DD.MM.YYYY»  
- Кнопка **Оплатить подписку** *или* **Отменить подписку**  
- Текст ошибки оплаты

**При нажатии**

| Элемент | Поведение |
|---------|-----------|
| Оплатить | `POST /payments` → `confirmation_url` → внешний браузер (ЮKassa) |
| После оплаты | Webhook → БД subscription ACTIVE → при возврате в app reload |
| Отменить | Диалог подтверждения → cancel → статус обновлён; доступ до конца оплаченного периода (**SaaS target**) |
| Назад | Home перечитывает subscription |

**SaaS target:** тарифы (месяц/год), чеки/история платежей, письмо на email, grace period, retry failed payment, deep-link return в приложение.

---

### 4.5 Настройки (ТЗ §20)

**Секция «Подключение»**

| Переключатель | ON | OFF |
|---------------|----|-----|
| **Автозапуск** | BootReceiver может поднять сервис (Android) | Только ручной старт |
| **Автоподключение** | После login/bootstrap при ACTIVE — Connect | Только по орбу |
| **Автовыбор Relay** | pickBestRelay по RTT/health | Учитывать preferred region/server |

**Секция «Маршрутизация»**

| Пункт | Что открывает | Данные на пункте |
|-------|---------------|------------------|
| **Исключения** | Экран доменов DIRECT | Число доменов / «синхронизировано» |
| **Приложения без VPN** | App bypass | Число пакетов |

**Секция «Поддержка»**

| Пункт | Действие |
|-------|----------|
| **Диагностика** | Экран логов и статуса |

**SaaS target:** Выйти из аккаунта; Удалить аккаунт; Язык; Тема; Уведомления о сбоях; «О приложении» (версия, build, connectFlow).

---

### 4.6 Исключения (User Rules / DIRECT)

**Данные:** поле домена (`*.mybank.ru`), список, пустое состояние, валидация.

| Действие | Поведение |
|----------|-----------|
| Добавить | Валидация → список → save local + `PUT /exclusions` → hot-reload rules если VPN up |
| Удалить | Убрать из списка → sync |
| Назад | Вернуть список в Settings |

Правило: пользовательские исключения **важнее** серверных domain rules для этого домена (DIRECT).

---

### 4.7 Приложения без VPN (App bypass)

**Зачем (не путать с DIRECT):** только `addDisallowedApplication` снимает флаг VPN у приложения. Нужно для банков/Госуслуг, которые блокируют при активном VPN-профиле.

**Данные:** поиск, чекбоксы (название + package), предустановленные эвристики (банки, госуслуги, ФНС…).

| Действие | Поведение |
|----------|-----------|
| Чекбокс | Локальный выбор |
| Готово | Сохранить packages; SnackBar «переподключите VPN» |
| Следующий Connect | Пакеты в bypass списка VpnService |

---

### 4.8 Диагностика

**Данные (строки)**  
Статус VPN · Relay · RTT · Время соединения · Код ошибки · Версия клиента · ОС · Connect log (на устройстве).

В логе допустимо: relay id/host/port, HTTP status, auth codes, build label, события `udp/443` / `tcp/8443`.  
Запрещено: URL сайтов, payload, `connection_config` целиком.

| Действие | Поведение |
|----------|-----------|
| Обновить | Считать native status + log |
| Копировать | Clipboard для поддержки |
| Очистить | Очистить Flutter + native log |

---

### 4.9 Статистика

**Сейчас:** заглушка «в разработке».

**SaaS target (ТЗ telemetry → UI):**
- Время online за день/неделю  
- Средний RTT, число reconnect  
- Доля трафика DIRECT vs RELAY (агрегаты без URL)  
- Последние ошибки (коды)  
- Кнопка «Обновить»

---

## 5. Состояния орба / VPN (матрица)

| Состояние UI | Условие | Текст / поведение |
|--------------|---------|-------------------|
| Idle | Не подключено | «Авто-маршрут готов», орб ждёт tap |
| Permission | Первый Connect | Системный диалог Android VPN |
| Connecting | Handshake | Анимация; лог PrepareRelay / candidate |
| Connected | Tunnel up | «Система активна», timer, ping |
| Degraded | Relay bad | Автосмена relay (**SaaS/ТЗ**); краткий toast |
| Error | Handshake fail / empty config | Сообщение + возможность Retry |
| Blocked (billing) | Нет ACTIVE | Не стартует tunnel |

Целевые SLA (ТЗ §22): connect ≤ 5 с; recover ≤ 10 с; cold start ≤ 2 с.

---

## 6. Фоновые сценарии системы

### 6.1 Выбор маршрута (каждый поток)

```
Запрос → exclusions? → domain/CIDR rules → DefaultMode (DIRECT)
         ├─ DIRECT  → вне relay (и по возможности вне TUN / bypass)
         ├─ RELAY   → Hysteria2
         └─ FALLBACK→ повтор/другой путь при сбое
```

### 6.2 Fallback портов (ТЗ §10 + ADR-014)

Порядок клиента: **UDP 443 → 8443 → 24443 → TCP underlay 8443 → 24443**  
(TCP/443 на prod занят Caddy — осознанное отклонение от буквы ТЗ.)

Пользователь **не выбирает порт**; в диагностике видит `hysteria ok via udp/443` или `tcp/8443`.

### 6.3 Обновления (ТЗ §16)

| Что | Как | UI |
|-----|-----|-----|
| Правила | Poll `GET /rules` + hot-reload | Без UI; версия в логе |
| Config | `GET /config` | Диалог soft/hard update |
| Relay list | Poll / при открытии Servers | Обновление ping/health |
| Клиент APK | `latest_client_version` + download URL | Диалог «Скачать» |
| Сертификаты | Caddy / LE на сервере | Невидимо |

### 6.4 Сессия / токены

- Access JWT короткий; refresh ротация  
- Secure storage (+ миграция со старых prefs)  
- 401 / auth-expired на API → refresh → fail → Onboarding  
- Logout (**SaaS target** на Settings): clear tokens → Onboarding  

---

## 7. Admin Panel (оператор SaaS)

URL: `/admin/` · Auth: `X-Admin-Key`

### 7.1 Вкладка Health
- Проверка доступности admin API (`GET /servers/all`)  
- Индикатор: backend жив / ошибка ключа  

### 7.2 Вкладка Users
**Таблица:** email · статус подписки · active_until · created · user id  

| Действие | Поведение |
|----------|-----------|
| Refresh | Перечитать список |
| **SaaS target** | Выдать/забрать Premium вручную, бан, сброс сессий, поиск |

### 7.3 Вкладка Relays
**Таблица:** id · region · host:port · healthy · RTT · load  

| Действие | Поведение |
|----------|-----------|
| Register form | Создать relay + `connection_config` |
| Delete | Удалить из реестра |
| Refresh | Актуальный health от healthmonitor |

### 7.4 Вкладка Rules
- JSON редактор rule set (kind/pattern/mode)  
- **Publish** → новая version → клиенты подхватывают poll’ом  

### 7.5 Вкладка Config
Поля: `min_supported_client_version`, `latest_client_version`, `client_download_url`, telemetry on/off, poll intervals.  
**Publish** → клиенты видят update dialog / меняют интервалы.

### 7.6 Ops вне UI (но часть SaaS)
- Grafana/Prometheus: CPU, RAM, relay load, RTT, errors  
- Postgres backup + off-site encrypt  
- Healthmonitor: TCP probe → `POST /servers/health`  

---

## 8. Платформы (ТЗ §3) — единый сценарий, разные адаптеры

| Платформа | Сетевой слой | UI-сценарий |
|-----------|--------------|-------------|
| Android 10+ | VpnService | Реализован (MVP) |
| Windows 10/11 | WFP | Тот же Home/Connect/Servers/Settings (**open**) |
| macOS 13+ | Network Extension | То же (**open**) |
| iOS 17+ | Packet Tunnel | То же + App Store / VPN profile UX (**open**) |

Бизнес-логика (auth, rules, billing, decision) — общая (≥90% core).

---

## 9. Негативные и edge-сценарии

| Ситуация | Ожидание продукта |
|----------|-------------------|
| Нет интернета при Connect | Error + Retry; без зависания UI |
| UDP блокирован оператором | Тихий fallback на TCP underlay |
| Relay unhealthy | Не выбирать в auto; переключить при degrade |
| Истёк access, жив refresh | Тихий refresh, Connect продолжается |
| Истёк refresh | Экран входа, сообщение «сессия истекла» |
| Hard update (version < min) | Блокирующий диалог без «Позже» |
| Банк «обнаружил VPN» | App в bypass → reconnect |
| Пользователь добавил `*.ru` в exclusions | Уже DIRECT; без вреда |
| Отмена подписки mid-period | SaaS: работать до `active_until` |
| Два устройства | SaaS target: лимит устройств / список устройств в аккаунте |
| Root / эмулятор | Не блокер MVP; политика на усмотрение |

---

## 10. Чеклист приёмки «как SaaS» (по ТЗ §22 + расширение)

### Обязательный MVP (из ТЗ)
- [x] Регистрация / логин  
- [ ] Оплата и активация подписки **live ЮKassa** (код есть, live — skipped)  
- [x] Connect одной кнопкой  
- [x] Маршрутизация по правилам  
- [x] Зарубежное через relay / RU напрямую (+ bypass)  
- [x] Fallback портов / TCP underlay  
- [x] Обновление rules/config  
- [ ] Автосмена relay при деградации — довести до наблюдаемого UX  
- [ ] SLA timing формально замерить на device  

### Зрелый SaaS (сверх MVP, но нужен для «полноценного продукта»)
- [ ] Password reset, logout, delete account  
- [ ] Статистика с реальными метриками  
- [ ] История платежей / счета  
- [ ] Multi-device policy  
- [ ] Support channel (email/Telegram) из Диагностики  
- [ ] Брендовый домен вместо nip.io  
- [ ] Windows / iOS / macOS клиенты  
- [ ] Admin: ручное управление подпиской, аудит действий  
- [ ] Статус-страница (status.streampass…)  
- [ ] Онбординг с 2–3 экранами «зачем / как / privacy»  

---

## 11. Скрипт демо для инвестора / QA (5 минут)

1. Открыть Onboarding → зарегистрировать тестовый email.  
2. Home → баннер подписки → (на staging) активировать Premium.  
3. Connect → разрешить VPN → дождаться Connected + ping.  
4. YouTube — играет; Госуслуги/банк — без «отключите VPN».  
5. Серверы → выбрать другой регион → (reconnect) ping обновился.  
6. Настройки → Диагностика → в логе `hysteria ok via …` → Скопировать.  
7. Disconnect → статус idle.  
8. Admin `/admin/` → Users видит аккаунт; Rules version; Relays healthy.

---

## 12. Словарь для UI-копирайта

| Не говорить | Говорить |
|-------------|----------|
| Ускорение интернета | Стабильный доступ к зарубежным сервисам |
| VPN (в маркетинге) | Умная маршрутизация / ускоритель |
| Мы видим ваши сайты | Только техническая телеметрия, без истории |

---

*Документ для Product / QA / Design / Dev. Источник истины по «должен» — ТЗ; по «есть сейчас» — `03_CurrentState.md` и код клиента.*
