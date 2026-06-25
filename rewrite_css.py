import re

css = """
:root {
    /* Material Design 3 Dark Theme (Tonal Palette based on Blue) */
    --md-sys-color-primary: #a8c7fa;
    --md-sys-color-on-primary: #062e6f;
    --md-sys-color-primary-container: #0842a0;
    --md-sys-color-on-primary-container: #d3e3fd;
    
    --md-sys-color-secondary: #c2c7cf;
    --md-sys-color-on-secondary: #2b313a;
    --md-sys-color-secondary-container: #414751;
    --md-sys-color-on-secondary-container: #dee3eb;

    --md-sys-color-error: #ffb4ab;
    --md-sys-color-error-container: #93000a;
    --md-sys-color-on-error: #690005;
    --md-sys-color-on-error-container: #ffdad6;

    --md-sys-color-success: #6dd58c;
    --md-sys-color-success-container: #00521c;
    --md-sys-color-on-success-container: #8bf2a6;

    --md-sys-color-warning: #eaddff;
    --md-sys-color-warning-container: #5b467a;
    --md-sys-color-on-warning-container: #f5eeff;
    
    --md-sys-color-background: #111318;
    --md-sys-color-on-background: #e2e2e9;
    
    --md-sys-color-surface: #111318;
    --md-sys-color-on-surface: #e2e2e9;
    --md-sys-color-surface-variant: #44474e;
    --md-sys-color-on-surface-variant: #c4c6d0;
    
    --md-sys-color-surface-container-lowest: #0c0e13;
    --md-sys-color-surface-container-low: #191c20;
    --md-sys-color-surface-container: #1d2024;
    --md-sys-color-surface-container-high: #282a2f;
    --md-sys-color-surface-container-highest: #33353a;

    --md-sys-color-outline: #8e9099;
    --md-sys-color-outline-variant: #44474e;
    
    --md-sys-color-disabled-bg: rgba(226, 226, 233, 0.12);
    --md-sys-color-disabled-text: rgba(226, 226, 233, 0.38);
    
    /* Elevations */
    --md-sys-elevation-1: 0px 1px 2px 0px rgba(0, 0, 0, 0.3), 0px 1px 3px 1px rgba(0, 0, 0, 0.15);
    --md-sys-elevation-2: 0px 1px 2px 0px rgba(0, 0, 0, 0.3), 0px 2px 6px 2px rgba(0, 0, 0, 0.15);
}

:root.light-theme {
    --md-sys-color-primary: #0842a0;
    --md-sys-color-on-primary: #ffffff;
    --md-sys-color-primary-container: #d3e3fd;
    --md-sys-color-on-primary-container: #001c3b;
    
    --md-sys-color-secondary: #535f70;
    --md-sys-color-on-secondary: #ffffff;
    --md-sys-color-secondary-container: #d7e3f8;
    --md-sys-color-on-secondary-container: #101c2b;

    --md-sys-color-error: #ba1a1a;
    --md-sys-color-error-container: #ffdad6;
    --md-sys-color-on-error: #ffffff;
    --md-sys-color-on-error-container: #410002;

    --md-sys-color-success: #146c2e;
    --md-sys-color-success-container: #a0f9b9;
    --md-sys-color-on-success-container: #002106;

    --md-sys-color-warning: #6d5e8f;
    --md-sys-color-warning-container: #f5eeff;
    --md-sys-color-on-warning-container: #271a47;
    
    --md-sys-color-background: #fdfbff;
    --md-sys-color-on-background: #1a1c1e;
    
    --md-sys-color-surface: #fdfbff;
    --md-sys-color-on-surface: #1a1c1e;
    --md-sys-color-surface-variant: #dfe2eb;
    --md-sys-color-on-surface-variant: #44474e;
    
    --md-sys-color-surface-container-lowest: #ffffff;
    --md-sys-color-surface-container-low: #f4f3f7;
    --md-sys-color-surface-container: #eeedf1;
    --md-sys-color-surface-container-high: #e8e7eb;
    --md-sys-color-surface-container-highest: #e2e2e5;

    --md-sys-color-outline: #74777f;
    --md-sys-color-outline-variant: #c4c6d0;
    
    --md-sys-color-disabled-bg: rgba(26, 28, 30, 0.12);
    --md-sys-color-disabled-text: rgba(26, 28, 30, 0.38);
    
    --md-sys-elevation-1: 0px 1px 2px 0px rgba(0, 0, 0, 0.3), 0px 1px 3px 1px rgba(0, 0, 0, 0.15);
    --md-sys-elevation-2: 0px 1px 2px 0px rgba(0, 0, 0, 0.3), 0px 2px 6px 2px rgba(0, 0, 0, 0.15);
}

body {
    background-color: var(--md-sys-color-background);
    color: var(--md-sys-color-on-background);
    font-family: 'Roboto', sans-serif;
    margin: 0;
    padding: 0;
    -webkit-font-smoothing: antialiased;
}

/* Material Icons */
.material-symbols-outlined {
    vertical-align: middle;
    font-variation-settings: 'FILL' 1, 'wght' 400, 'GRAD' 0, 'opsz' 24;
}

/* Typography */
.md3-title-large { font-size: 22px; font-weight: 400; line-height: 28px; letter-spacing: 0px; }
.md3-title-medium { font-size: 16px; font-weight: 500; line-height: 24px; letter-spacing: 0.15px; }
.md3-body-large { font-size: 16px; font-weight: 400; line-height: 24px; letter-spacing: 0.5px; }
.md3-body-medium { font-size: 14px; font-weight: 400; line-height: 20px; letter-spacing: 0.25px; }
.md3-label-large { font-size: 14px; font-weight: 500; line-height: 20px; letter-spacing: 0.1px; }
.md3-label-medium { font-size: 12px; font-weight: 500; line-height: 16px; letter-spacing: 0.5px; }

/* Top App Bar */
.top-app-bar {
    background-color: var(--md-sys-color-surface-container);
    padding: 16px 24px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    position: sticky;
    top: 0;
    z-index: 100;
}

.app-title-icon {
    color: var(--md-sys-color-primary);
}

/* Container */
.container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 24px;
}

/* ====== BUTTONS ====== */
.btn {
    background-color: var(--md-sys-color-primary);
    color: var(--md-sys-color-on-primary);
    border: none;
    padding: 10px 24px;
    border-radius: 9999px; /* MD3 Pill */
    font-family: 'Roboto', sans-serif;
    font-size: 14px;
    font-weight: 500;
    line-height: 20px;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    position: relative;
    overflow: hidden;
    transition: background-color 0.2s, box-shadow 0.2s, color 0.2s;
}

/* State Layer overlay */
.btn::after {
    content: '';
    position: absolute;
    inset: 0;
    background-color: var(--md-sys-color-on-primary);
    opacity: 0;
    transition: opacity 0.2s;
}

.btn:hover { box-shadow: var(--md-sys-elevation-1); }
.btn:hover::after { opacity: 0.08; }
.btn:active::after { opacity: 0.12; }

/* Secondary Button (Tonal) */
.btn.secondary {
    background-color: var(--md-sys-color-secondary-container);
    color: var(--md-sys-color-on-secondary-container);
}
.btn.secondary::after {
    background-color: var(--md-sys-color-on-secondary-container);
}

.btn:disabled {
    background-color: var(--md-sys-color-disabled-bg);
    color: var(--md-sys-color-disabled-text);
    cursor: not-allowed;
    box-shadow: none;
}
.btn:disabled::after { display: none; }

/* Icon Button */
.icon-btn {
    width: 40px;
    height: 40px;
    padding: 0;
    border-radius: 50%;
    justify-content: center;
    background: transparent;
    color: var(--md-sys-color-on-surface-variant);
}
.icon-btn::after {
    background-color: var(--md-sys-color-on-surface-variant);
}
.icon-btn:hover { box-shadow: none; }
.icon-btn.active { color: var(--md-sys-color-primary); }

/* Segmented Button */
.segmented-button {
    display: inline-flex;
    border: 1px solid var(--md-sys-color-outline);
    border-radius: 9999px; /* Full pill wrapper */
    overflow: hidden;
    margin-bottom: 24px;
    height: 40px;
}
.seg-btn {
    background: transparent;
    color: var(--md-sys-color-on-surface);
    border: none;
    padding: 0 16px;
    font-family: 'Roboto', sans-serif;
    font-weight: 500;
    font-size: 14px;
    cursor: pointer;
    border-right: 1px solid var(--md-sys-color-outline);
    display: inline-flex;
    align-items: center;
    gap: 8px;
    position: relative;
    overflow: hidden;
}
.seg-btn:last-child { border-right: none; }

.seg-btn::after {
    content: '';
    position: absolute;
    inset: 0;
    background-color: var(--md-sys-color-on-surface);
    opacity: 0;
    transition: opacity 0.2s;
}
.seg-btn:hover::after { opacity: 0.08; }
.seg-btn:active::after { opacity: 0.12; }

.seg-btn.active {
    background-color: var(--md-sys-color-secondary-container);
    color: var(--md-sys-color-on-secondary-container);
}
.seg-btn.active::after {
    background-color: var(--md-sys-color-on-secondary-container);
}


/* Spin Animation */
.spin { animation: spin 2s linear infinite; }
@keyframes spin { 100% { transform: rotate(360deg); } }

.tab-content { display: none; }
.tab-content.active { display: block; animation: fadeIn 0.2s ease; }
@keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }

/* ====== CARDS ====== */
/* Elevated Card */
.card {
    background-color: var(--md-sys-color-surface-container-low);
    border-radius: 16px; 
    box-shadow: var(--md-sys-elevation-1);
    padding: 24px;
    margin-bottom: 24px;
}

/* Filled Card (For groups) */
.group-card {
    background-color: var(--md-sys-color-surface-container-highest);
    border-radius: 16px;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 12px;
}


/* Grid */
.grid-groups {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
    margin-bottom: 32px;
}

/* ====== FORMS & INPUTS ====== */
/* Checkbox MD3 */
label.md3-checkbox-label {
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    color: var(--md-sys-color-on-surface);
}
input[type="checkbox"] {
    appearance: none;
    background-color: transparent;
    margin: 0;
    font: inherit;
    color: currentColor;
    width: 18px;
    height: 18px;
    border: 2px solid var(--md-sys-color-on-surface-variant);
    border-radius: 2px;
    display: grid;
    place-content: center;
    transition: all 0.2s ease-in-out;
}
input[type="checkbox"]::before {
    content: "";
    width: 10px;
    height: 10px;
    transform: scale(0);
    transition: 120ms transform ease-in-out;
    box-shadow: inset 1em 1em var(--md-sys-color-on-primary);
    background-color: var(--md-sys-color-on-primary);
    transform-origin: center;
    clip-path: polygon(14% 44%, 0 65%, 50% 100%, 100% 16%, 80% 0%, 43% 62%);
}
input[type="checkbox"]:checked {
    background-color: var(--md-sys-color-primary);
    border-color: var(--md-sys-color-primary);
}
input[type="checkbox"]:checked::before {
    transform: scale(1);
}

/* Select MD3 Filled style */
select.md3-select {
    appearance: none;
    background-color: var(--md-sys-color-surface-variant);
    color: var(--md-sys-color-on-surface);
    border: none;
    border-bottom: 1px solid var(--md-sys-color-on-surface-variant);
    border-radius: 4px 4px 0 0;
    padding: 8px 32px 8px 16px;
    font-size: 14px;
    font-family: inherit;
    outline: none;
    transition: border-bottom-color 0.2s, background-color 0.2s;
    background-image: url('data:image/svg+xml;utf8,<svg fill="%23c4c6d0" height="24" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg"><path d="M7 10l5 5 5-5z"/></svg>');
    background-repeat: no-repeat;
    background-position: right 8px center;
}
.light-theme select.md3-select {
    background-image: url('data:image/svg+xml;utf8,<svg fill="%2344474e" height="24" viewBox="0 0 24 24" width="24" xmlns="http://www.w3.org/2000/svg"><path d="M7 10l5 5 5-5z"/></svg>');
}
select.md3-select:focus {
    border-bottom: 2px solid var(--md-sys-color-primary);
    background-color: color-mix(in srgb, var(--md-sys-color-surface-variant) 80%, var(--md-sys-color-primary) 20%);
}

/* ====== TABLE ====== */
.table-container {
    border: 1px solid var(--md-sys-color-outline-variant);
    border-radius: 12px;
    overflow: hidden;
    background-color: var(--md-sys-color-surface);
}

table {
    width: 100%;
    border-collapse: collapse;
    text-align: left;
}

th {
    background-color: var(--md-sys-color-surface);
    color: var(--md-sys-color-on-surface-variant);
    font-size: 14px;
    font-weight: 500;
    padding: 16px;
    border-bottom: 1px solid var(--md-sys-color-outline-variant);
}

td {
    padding: 16px;
    border-bottom: 1px solid var(--md-sys-color-outline-variant);
    font-size: 14px;
    color: var(--md-sys-color-on-surface);
}

tr:last-child td { border-bottom: none; }

.node-row {
    cursor: pointer;
    transition: background-color 0.2s;
}

.node-row:hover {
    background-color: var(--md-sys-color-surface-container-high);
}

/* ====== BADGES ====== */
.badge {
    display: inline-flex;
    align-items: center;
    padding: 4px 8px;
    border-radius: 8px;
    font-size: 12px;
    font-weight: 500;
    gap: 4px;
}

.badge.primary { background-color: var(--md-sys-color-primary-container); color: var(--md-sys-color-on-primary-container); }
.badge.success { background-color: var(--md-sys-color-success-container); color: var(--md-sys-color-on-success-container); }
.badge.error { background-color: var(--md-sys-color-error-container); color: var(--md-sys-color-on-error-container); }
.badge.warning { background-color: var(--md-sys-color-warning-container); color: var(--md-sys-color-on-warning-container); }

/* Score Box */
.score-box {
    font-size: 16px;
    font-weight: 500;
    color: var(--md-sys-color-on-surface);
    padding: 4px 12px;
    border-radius: 12px;
    background-color: var(--md-sys-color-surface-container-highest);
    display: inline-block;
}

/* Flash Effects */
.flash-green { animation: flashGreen 1s ease-out; }
.flash-red { animation: flashRed 1s ease-out; }

@keyframes flashGreen {
    0% { background-color: var(--md-sys-color-success); color: #000; }
    100% { background-color: var(--md-sys-color-surface-container-highest); }
}
@keyframes flashRed {
    0% { background-color: var(--md-sys-color-error); color: #000; }
    100% { background-color: var(--md-sys-color-surface-container-highest); }
}

/* Logs Console */
.console {
    background-color: #000;
    color: #e2e2e9;
    font-family: 'Roboto Mono', monospace;
    padding: 16px;
    border-radius: 12px;
    height: 500px;
    overflow-y: auto;
    font-size: 13px;
    line-height: 1.6;
}

.log-line {
    display: flex;
    margin-bottom: 4px;
    animation: slideIn 0.3s ease-out;
}
@keyframes slideIn { from{transform: translateX(-10px); opacity:0;} to{transform: translateX(0); opacity:1;} }

.log-time { color: var(--md-sys-color-outline); margin-right: 12px; }
.log-badge { padding: 2px 6px; border-radius: 4px; font-size: 11px; margin-right: 12px; font-weight: bold; }
.log-msg { word-break: break-all; }

.log-info .log-badge { background: var(--md-sys-color-primary-container); color: var(--md-sys-color-on-primary-container); }
.log-success .log-badge { background: var(--md-sys-color-success-container); color: var(--md-sys-color-on-success-container); }
.log-warning .log-badge { background: var(--md-sys-color-warning-container); color: var(--md-sys-color-on-warning-container); }
.log-error .log-badge { background: var(--md-sys-color-error-container); color: var(--md-sys-color-on-error-container); }

.log-info .log-msg { color: #e2e2e9; }
.log-success .log-msg { color: #6dd58c; }
.log-warning .log-msg { color: #eaddff; }
.log-error .log-msg { color: #ffb4ab; }
"""

with open("frontend/src/index.css", "w", encoding="utf-8") as f:
    f.write(css)

print("index.css successfully rewritten with MD3 specs.")
