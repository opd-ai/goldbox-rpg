// Package main provides an Android-embeddable Gold Box RPG server.
//
// This command implements the full Gold Box RPG Engine server designed to be
// compiled for Android ARM64 and packaged as a native library in the companion
// Android application. It hosts the same WASM-based game interface as the
// regular server (cmd/server), including the Ebitengine client, JSON-RPC
// backend, and WebSocket real-time communication.
//
// The binary is renamed to libwebservice.so and placed in the jniLibs
// directory so it receives the correct SELinux context (apk_data_file) and
// can be executed on production devices with locked bootloaders.
//
// # Endpoints
//
// The server exposes the same endpoints as the main Gold Box RPG server:
//
//   - GET / — serves the WASM splash screen that loads the Ebitengine client
//   - POST / — JSON-RPC 2.0 API for game actions
//   - WebSocket — real-time game event updates
//   - GET /health, /ready, /live — observability endpoints
//
// # Building for Android
//
// Use the provided build.sh script to cross-compile for ARM64 and prepare
// bundled assets (WASM binary, web UI, game data):
//
//	cd cmd/android-service
//	./build.sh
//
// The compiled binary is placed in android/app/src/main/jniLibs/arm64-v8a/libwebservice.so
// and is automatically extracted by the Android package manager into the app's
// native library directory at install time. At runtime the binary path is
// resolved via Context.getApplicationInfo().nativeLibraryDir.
//
// Web and game data assets are bundled into android/app/src/main/assets/ and
// extracted to internal storage by the Android application at first launch.
// The Go process receives the paths via WEB_DIR and DATA_DIR environment
// variables.
//
// # Android Application
//
// The android/ subdirectory contains a complete Android Studio project with:
//
//   - MainActivity.kt — UI, asset extraction, and Go process lifecycle management
//   - activity_main.xml — layout with service controls and log viewer
//   - AndroidManifest.xml — permissions for INTERNET and ACCESS_NETWORK_STATE
//   - build.gradle — app-level Gradle build configuration targeting API 24+
package main
