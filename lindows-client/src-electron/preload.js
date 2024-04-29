const { BrowserWindow, contextBridge, ipcRenderer, clipboard } = require('electron');

contextBridge.exposeInMainWorld('clipboard', {
    writeText: (text) => clipboard.writeText(text),
    readText: () => clipboard.readText()
});

contextBridge.exposeInMainWorld('electron', {
    startDrag: (fileName) => {
        ipcRenderer.send('ondragstart', path.join(process.cwd(), fileName))
    }
});

contextBridge.exposeInMainWorld('electronWindow', {
    // 创建窗口, 返回窗口的id(number 类型)
    createWindow: (width, height) => ipcRenderer.invoke('create-window', width, height),
    // 最小化窗口
    minimizeWindow: (windowId) => ipcRenderer.send('minimize-window', windowId),
    // 最大化窗口
    maximizeWindow: (windowId) => ipcRenderer.send('maximize-window', windowId),
    // 还原窗口
    unmaximizeWindow: (windowId) => ipcRenderer.send('unmaximize-window', windowId),
    // 关闭窗口
    closeWindow: (windowId) => ipcRenderer.send('close-window', windowId),

    // 检查当前窗口是否是主窗口
    isMainWindow: () => ipcRenderer.invoke('is-main-window'),
    // 获取主窗口的id
    getMainWindowId: () => ipcRenderer.invoke('get-main-window-id'),
});