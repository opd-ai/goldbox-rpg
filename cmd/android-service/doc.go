// Package main provides an Android-embeddable Go HTTP web service.
//
// This command implements a lightweight HTTP server designed to be compiled for
// Android ARM64 and packaged as a native library in the companion Android
// application. The binary is renamed to libwebservice.so and placed in the
// jniLibs directory so it receives the correct SELinux context (apk_data_file)
// and can be executed on production devices with locked bootloaders.
//
// The service binds to all network interfaces (0.0.0.0) so it is accessible
// from both localhost (127.0.0.1) and the device's LAN IP address.
//
// # Endpoints
//
//   - GET / — returns "Hello from Go service"
//   - GET /status — returns JSON with server status, timestamp, and hostname
//   - GET /ip — returns the detected LAN IP address as plain text
//
// # Building for Android
//
// Use the provided build.sh script to cross-compile for ARM64:
//
//	cd cmd/android-service
//	./build.sh
//
// The compiled binary is placed in android/app/src/main/jniLibs/arm64-v8a/libwebservice.so
// and is automatically extracted by the Android package manager into the app's
// native library directory at install time. At runtime the binary path is
// resolved via Context.getApplicationInfo().nativeLibraryDir.
//
// # Android Application
//
// The android/ subdirectory contains a complete Android Studio project with:
//
//   - MainActivity.kt — UI and Go process lifecycle management
//   - activity_main.xml — layout with service controls and log viewer
//   - AndroidManifest.xml — permissions for INTERNET and ACCESS_NETWORK_STATE
//   - build.gradle — app-level Gradle build configuration targeting API 24+
package main
