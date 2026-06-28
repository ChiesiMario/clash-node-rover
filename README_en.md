[English](README_en.md) | [繁體中文](README.md) | [简体中文](README_zh-CN.md)

# Clash Node Rover

<p align="center">
  <img src="Image/1.png" alt="Clash Node Rover Dashboard" width="100%" />
</p>

**Clash Node Rover** is a high-performance background engine and graphical management tool built with Rust and Tauri, specifically designed to monitor and automatically switch your Clash / Clash Meta proxies to the fastest and most stable node available.

---

## ✨ Key Features

- 🚀 **Smart Dynamic Speed Engine**: Concurrently tests your proxy nodes in the background with minimal resource usage. Using a weighted average and jitter to calculate a comprehensive score, it seamlessly and automatically switches nodes when a significantly better one is found.
- 📊 **Historical Performance Tracking & Charts**: Say goodbye to point-in-time speed tests! We exclusively provide up to 7 days of historical latency charts, giving you a clear view of node stability across different times.
- 🛡️ **Local Network Check**: Before executing any speed test or node switch, the engine automatically detects local network connectivity (via DNS probing) to prevent misjudging all high-quality nodes as dead due to a local outage.
- 🌐 **Deep HTTP Connectivity Verification**: Supports real-world web access testing (via specified HTTP Proxies) before switching, ensuring the node doesn't just "Ping", but can smoothly access the internet.
- 🎨 **Modern Multilingual Interface**: Built with React + TailwindCSS featuring a beautiful Glassmorphism UI. Supports English, Traditional Chinese, and Simplified Chinese.

---

## 📸 Interface Preview

### Dashboard
Monitor system status, API connectivity, active groups, and current speed test progress in real time.

<img src="Image/1.png" alt="Dashboard" width="100%" />

### Node Ranking & History
Sorts all nodes by score, supports region filtering, and allows clicking on a node to expand its 24H / 3-Day / 7-Day historical latency line chart.

<img src="Image/2.png" alt="Node Ranking" width="100%" />

### Advanced Settings
Highly customizable settings for Tolerance, Backoff Rounds, and maximum concurrency to suit various network environments.

<img src="Image/3.png" alt="Settings" width="100%" />

---

## 📥 Installation & Usage

### 1. Download & Install
Head over to the [Releases](../../releases) page to download the latest `.exe` or `.msi` package.

### 2. Initial Setup Wizard
1. **Connect API**: Upon launch, the software will guide you to connect to the Clash API (usually `http://127.0.0.1:9090`). If you have a Secret configured, enter it as well.
2. **Select Groups**: Check the **Proxy Groups (Selectors)** you want Rover to monitor and automatically switch.
3. **Finish**: Start enjoying a fully automated network optimization experience!

---

## ⚙️ Advanced Core Mechanics

### Switch Tolerance
To avoid frequent node switching that disrupts network connections (e.g., buffering while watching videos or gaming lag), Rover introduces a tolerance mechanism. The system will maintain the current node if the new node's comprehensive score does not exceed the "current node's score + tolerance".

### Backoff Algorithm
For nodes that consecutively time out during speed tests or fail the real-world web test, the system will mark them as in a "Backoff State" for a set number of rounds. Nodes in backoff will be skipped in subsequent speed tests, significantly reducing the system's burden of sending invalid requests.

---

## 🛠️ Development & Building

This project is built using `Tauri v2`, `Rust`, and `React`. If you wish to compile it yourself or contribute:

```bash
# 1. Install frontend dependencies
npm install

# 2. Start developer mode
npm run dev

# 3. Compile for production (generates .exe and .msi on Windows)
npm run tauri build
```

## 📄 License
MIT License
