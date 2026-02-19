/**
 * native-bridge/bridge.js
 *
 * Injected via runtime.WindowExecJS on OnDomReady. Overrides the three
 * File System Access API entry points used by the TUIStudio web app:
 *
 *   window.showSaveFilePicker   → NativeSaveDialog  (writes file, returns path)
 *   window.showOpenFilePicker   → NativeOpenDialog  (returns file content string)
 *   window.showDirectoryPicker  → NativePickDirectory + NativeWriteFile
 *
 * Guard: window.go is only present inside the Wails webview. In a regular
 * browser this script is never injected, so native browser behaviour is kept.
 *
 * Fake handles implement only the subset of the File System Access API that
 * the web app actually calls (verified in fileOps.ts / downloadManager.ts).
 */

(function installNativeBridge() {
  if (typeof window.go === 'undefined') return;

  // Disable text selection globally — desktop apps should not behave like web pages.
  // Input elements are excluded so typing still works normally.
  const style = document.createElement('style');
  style.textContent = `
    *, *::before, *::after {
      -webkit-user-select: none;
      user-select: none;
    }
    input, textarea, [contenteditable] {
      -webkit-user-select: text;
      user-select: text;
    }
  `;
  document.head.appendChild(style);

  const go = window.go.main.App;

  /* ------------------------------------------------------------------ */
  /* showSaveFilePicker                                                   */
  /* fileOps.ts: fileHandle.createWritable() → writable.write(json) → writable.close() */
  /* ------------------------------------------------------------------ */
  window.showSaveFilePicker = async function (opts) {
    const suggestedName = (opts && opts.suggestedName) || 'untitled';

    // Buffer content from write(); flush to Go (which shows the dialog) on close().
    let _bufferedContent = '';

    const writable = {
      write(data) {
        _bufferedContent = data;
        return Promise.resolve();
      },
      close: async function () {
        const path = await go.NativeSaveDialog(suggestedName, _bufferedContent);
        if (path === '') {
          // Treat empty path as user cancel — mimic AbortError
          throw new DOMException('The user aborted a request.', 'AbortError');
        }
      },
    };

    const fileHandle = {
      createWritable: () => Promise.resolve(writable),
    };

    return fileHandle;
  };

  /* ------------------------------------------------------------------ */
  /* showOpenFilePicker                                                   */
  /* fileOps.ts: [fileHandle] → fileHandle.getFile() → file.text()      */
  /* ------------------------------------------------------------------ */
  window.showOpenFilePicker = async function (_opts) {
    const content = await go.NativeOpenDialog();

    if (content === '') {
      // User cancelled
      throw new DOMException('The user aborted a request.', 'AbortError');
    }

    const fakeFile = {
      text: () => Promise.resolve(content),
      arrayBuffer: () => Promise.resolve(new TextEncoder().encode(content).buffer),
    };

    const fileHandle = {
      getFile: () => Promise.resolve(fakeFile),
      kind: 'file',
      name: 'file.tui',
    };

    return [fileHandle];
  };

  /* ------------------------------------------------------------------ */
  /* showDirectoryPicker                                                  */
  /* downloadManager.ts:                                                  */
  /*   _directoryHandle = await showDirectoryPicker()                    */
  /*   _directoryHandle.name                                             */
  /*   _directoryHandle.getFileHandle(filename, {create:true})           */
  /*     → fileHandle.createWritable() → writable.write(content) → close */
  /* ------------------------------------------------------------------ */
  window.showDirectoryPicker = async function (_opts) {
    const dirPath = await go.NativePickDirectory();

    if (!dirPath || dirPath === '') {
      throw new DOMException('The user aborted a request.', 'AbortError');
    }

    // Extract folder name for display (matches _directoryHandle.name)
    const dirName = dirPath.replace(/\\/g, '/').split('/').filter(Boolean).pop() || dirPath;

    const dirHandle = {
      kind: 'directory',
      name: dirName,
      _path: dirPath,

      getFileHandle: function (filename, _opts) {
        const fullPath = dirPath.replace(/\\/g, '/').replace(/\/$/, '') + '/' + filename;

        let _bufferedContent = '';

        const writable = {
          write(data) {
            _bufferedContent = data;
            return Promise.resolve();
          },
          close: async function () {
            await go.NativeWriteFile(fullPath, _bufferedContent);
          },
        };

        const fileHandle = {
          kind: 'file',
          name: filename,
          createWritable: () => Promise.resolve(writable),
        };

        return Promise.resolve(fileHandle);
      },
    };

    return dirHandle;
  };

  console.debug('[TUIStudio] Native bridge installed — file dialogs route through Go IPC.');
})();
