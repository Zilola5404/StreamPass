# Интеграция общего Go-ядра в Android-клиент

Цель: переиспользовать существующий Go core (Decision Engine, Rule Engine,
Relay Manager) с бэкенда — как описано в ТЗ §4, доля общего кода должна
быть ≥90%.

## 1. Что нужно добавить в Go core

Если сейчас в Go core есть только *серверная* сторона Hysteria2 (relay),
для клиента нужен клиентский Hysteria2-стек. Варианты:

- использовать `github.com/apernet/hysteria` как библиотеку клиента
  (у него уже есть `core/client` пакет — не нужно писать транспорт заново,
  что совпадает с требованием ТЗ §8 "свой транспорт не разрабатывать");
- обернуть его функциями с плоской C-совместимой сигнатурой, которые
  понимает `gomobile bind` (только примитивные типы и интерфейсы
  в аргументах/возврате — структуры Go напрямую не экспортируются).

Пример точки входа (`go_core/mobile/tunnel.go`):

```go
package mobile

// StatusCallback — интерфейс, который Kotlin реализует на своей стороне;
// gomobile сгенерирует Java/Kotlin-обёртку под него.
type StatusCallback interface {
    OnConnecting()
    OnConnected(relay string, pingMs int)
    OnDisconnected()
    OnError(message string)
}

// StartTunnel запускается из StreamPassVpnService.establishTunnel().
// fd — файловый дескриптор TUN-интерфейса из VpnService.Builder().establish().
func StartTunnel(fd int, relayHost string, relayPort int, connectionConfig string, cb StatusCallback) {
    cb.OnConnecting()
    // ... hysteria2 client + sing-tun bridge on fd ...
}

func StopTunnel() {
    // закрыть соединение, освободить ресурсы
}
```

## 2. Сборка AAR

```bash
# Требуется JDK (javac) в PATH — например Android Studio JBR:
# export JAVA_HOME="/path/to/Android Studio/jbr"

go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest
gomobile init
cd client/go_core
go get -tool golang.org/x/mobile/cmd/gobind
gomobile bind -target=android -androidapi=21 -o streampasscore.aar ./mobile
cp streampasscore.aar ../android/app/libs/
```

Полученный `streampasscore.aar` кладётся в
`android/app/libs/streampasscore.aar` и подключается в
`android/app/build.gradle`:

```gradle
dependencies {
    implementation files('libs/streampasscore.aar')
}
```

## 3. Что заменить в StreamPassVpnService.kt

Раскомментировать импорт `streampasscore.Streampasscore` и заменить
стаб-блок в `establishTunnel()` реальным вызовом:

```kotlin
Streampasscore.startTunnel(
    fd.toLong(),
    relayHost,
    relayPort.toLong(),
    authPassword,
    object : Streampasscore.StatusCallback {
        override fun onConnecting() = emit("connecting")
        override fun onConnected(relay: String, pingMs: Long) =
            emit("connected", relay = relay, pingMs = pingMs.toInt())
        override fun onDisconnected() = emit("disconnected")
        override fun onError(message: String) = emit("error", error = message)
    }
)
```

После этого iOS-клиент сможет переиспользовать тот же `go_core/mobile`
пакет через `gomobile bind -target=ios`, а Windows/macOS — через обычную
Go-компиляцию в shared library (cgo), без Flutter-обёртки поверх сети —
только над UI.
