const { app, BrowserWindow, ipcMain, Menu } = require('electron');
const path = require('path');
const url = require('url')

let winIdCounter = 1;
let windowMap = new Map();

let mainWindow;

// const isDev = process.env.NODE_ENV !== 'production';
// TODO: 目前只能手动设置
const isDev = true;

const DEV_URL = 'http://localhost:1420';
const PROD_URL = url.format({
    pathname: path.join(__dirname, '..', 'dist', 'index.html'),
    protocol: 'file:',
    slashes: true
});

function createWindow() {
    // 创建浏览器窗口
    mainWindow = new BrowserWindow({
        frame: false,
        width: 800,
        height: 600,
        webPreferences: {
            nodeIntegration: true,
            devTools: true,
            preload: path.join(__dirname, 'preload.js')
        },
        backgroundColor: '#2e2c29',
    });

    mainWindow.once('ready-to-show', () => {
        mainWindow.show()
    })

    mainWindow.webContents.openDevTools();

    windowMap.set(winIdCounter, mainWindow);
    winIdCounter++;
    mainWindow.on('closed', () => {
        windowMap.delete(winIdCounter);
    });

    mainWindow.loadURL(isDev ? DEV_URL : PROD_URL);
}

// 隐藏菜单栏
Menu.setApplicationMenu(null)
// 创建自定义菜单
const menu = Menu.buildFromTemplate([
    {
        label: 'Developer',
        submenu: [
            {
                label: 'Toggle DevTools',
                accelerator: process.platform === 'darwin' ? 'Alt+Command+I' : 'Ctrl+Shift+I',
                click: function () {
                    mainWindow.webContents.toggleDevTools();
                }
            }
        ]
    }
]);

// 设置应用菜单
Menu.setApplicationMenu(menu);

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


//=============================================================//
// 窗口的最小化、最大化、关闭功能
//=============================================================//
ipcMain.on('minimize-window', (event, windowId) => {
    let win = windowMap.get(windowId);
    if (win) {
        win.minimize();
    }
});

ipcMain.on('maximize-window', (event, windowId) => {
    let win = windowMap.get(windowId);
    if (win) {
        win.maximize();
    }
});

ipcMain.on('unmaximize-window', (event, windowId) => {
    let win = windowMap.get(windowId);
    if (win) {
        win.unmaximize();
    }
});

ipcMain.on('close-window', (event, windowId) => {
    let win = windowMap.get(windowId);
    if (win) {
        win.close();
    }
});

//=============================================================//
// 创建子窗口
//=============================================================//
ipcMain.handle('create-window', async (event, width, height) => {
    let win = new BrowserWindow({
        width: width,
        height: height,
        webPreferences: {
            nodeIntegration: false,
            devTools: true,
            preload: path.join(__dirname, 'preload.js'),
            parent: mainWindow,
            additionalArguments: {
                additionalArguments: ['--window-type=sub-window'],
            },
        },
        backgroundColor: '#2e2c29',
    });

    windowMap.set(winIdCounter, win);
    winIdCounter++;
    win.on('closed', () => {
        windowMap.delete(winIdCounter);
    });

    subWindow.once('ready-to-show', () => {
        subWindow.show()
    })

    win.loadURL(isDev ? DEV_URL : PROD_URL);
    return win.id;
});

//=============================================================//
// 检查当前窗口是否是主窗口
//=============================================================//
ipcMain.handle('is-main-window', (event) => {
    return event.returnValue = mainWindow.id === mainWindow.id;
});

//=============================================================//
// 获取主窗口的id
//=============================================================//
ipcMain.handle('get-main-window-id', async (event) => {
    return mainWindow.id;
});