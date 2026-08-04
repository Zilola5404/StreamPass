# StreamPass Client (Flutter / Android)

Android-клиент StreamPass: UI на Flutter, туннель в Go (`go_core`) через gomobile AAR.

## Сборка APK (лёгкий arm64)

```bash
flutter pub get
flutter build apk --release --target-platform android-arm64
```

Файл: `build/app/outputs/flutter-apk/app-release.apk` (~18 MB).

## Go core (AAR)

```bash
# JAVA_HOME = Android Studio JBR
cd go_core
gomobile bind -target=android -androidapi=21 -o streampasscore.aar ./mobile
cp streampasscore.aar ../android/app/libs/
```

См. также `go_core/README.md`.

## Release signing (BL-013)

1. Скопируйте `android/key.properties.example` → `android/key.properties`
2. Создайте keystore:
   ```bash
   keytool -genkey -v -keystore android/upload-keystore.jks -keyalg RSA -keysize 2048 -validity 10000 -alias streampass
   ```
3. Заполните пароли в `key.properties` (файл в `.gitignore`)
4. `flutter build apk --release` подпишет production-ключом; без `key.properties` — debug (с предупреждением в логе Gradle).

## Основные экраны

Home (connect), servers, subscription, settings / exclusions, diagnostics, statistics.

API base задаётся в приложении (prod: `https://212-43-156-33.nip.io/api/v1`).
