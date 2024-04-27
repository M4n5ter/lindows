const { app, BrowserWindow } = require('electron');
const path = require('path');
const url = require('url')

function createWindow() {
    // 创建浏览器窗口
    let win = new BrowserWindow({
        width: 800,
        height: 600,
        webPreferences: {
            nodeIntegration: true,
            devTools: true
            // preload: path.join(__dirname, 'preload.js') // 如果需要预加载脚本
        }
    });

    // const isDev = process.env.NODE_ENV !== 'production';
    const isDev = false;

    const DEV_URL = 'http://localhost:1420';
    const PROD_URL = url.format({
        pathname: path.join(__dirname, 'dist', 'index.html'),
        protocol: 'file:',
        slashes: true
    });

    // 打开开发者工具
    win.webContents.openDevTools();

    win.loadURL(isDev ? DEV_URL : PROD_URL);
}

app.whenReady().then(createWindow);

app.on('window-all-closed', () => {
    if (process.platform !== 'darwin') {
        app.quit();
    }
});

app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
        createWindow();
    }
});
