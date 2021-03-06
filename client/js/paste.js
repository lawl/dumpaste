var G_dataTransferFile = null;

function getNewKey() {
    return window.crypto.subtle.generateKey({
            name: "AES-CBC",
            length: 256,
        },
        true,
        ["encrypt", "decrypt"]
    );
}

function getContentEncrypted(key, content) {
    return crypto.subtle.encrypt({
        name: 'AES-CBC',
        iv: new Uint8Array(16) // as we never reuse keys and always generate a fresh key, a 0bytes IV is fine
    }, key, content);
}

function decryptContent(key, content) {
    return crypto.subtle.decrypt({
        name: 'AES-CBC',
        iv: new Uint8Array(16) // as we never reuse keys and always generate a fresh key, 0bytes IV is fine
    }, key, content);
}

function pasteIt() {
    getNewKey().then(key => {
        var ulbox = document.getElementById('filef');
        var f = ulbox.files[0];

        if (!f && G_dataTransferFile) {
            f = G_dataTransferFile;
        }

        if (f) {
            readFile(f, function (data, fext) {
                var enccontent = getContentEncrypted(key, data);
                enccontent.then(blob => uploadContent(key, blob, f.type, fext));
            });
        } else {
            var enccontent = getContentEncrypted(key, new TextEncoder().encode(document.getElementById('r').value));
            enccontent.then(blob => uploadContent(key, blob, 'text/plain', 'file.txt'));
        }
    });
}

function uploadContent(key, blob, mime, fname) {

    var req = new XMLHttpRequest();
    req.open("POST", '/store?exp=' + document.getElementById('expires').value, true);
    req.onload = function (evt) {
        var res = req.responseText.split(/:/);
        if (res[0] == "OK") {
            document.getElementById('r').value = '';
            crypto.subtle.exportKey('raw', key).then(raw => window.location.hash = "#" + res[1] + ":" + arrayBufferToBase64(raw) + ':' + mime + ':' + fname);
        }
    };

    req.send(blob);
}

function findBaseURL() {
    return (window.location + '').split(/#/)[0];
}

function getBin(id, rawKey, mime) {
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4) {
            crypto.subtle.importKey('raw', base64ToArrayBuffer(rawKey), 'AES-CBC', true, ["encrypt", "decrypt"]).then(key =>
                decryptContent(key, xhr.response).then(data => showBin(data, mime))
            )
        }
    }
    xhr.open('GET', "/get/" + id, true);
    xhr.responseType = "arraybuffer";
    xhr.send('');
}

function showBin(c, mime) {
    var dataView = new DataView(c);
    var blob = new Blob([dataView], {
        type: mime
    });
    var blobURL = window.URL.createObjectURL(blob);
    document.getElementById('dlLink').href = blobURL;
    if (mime == 'text/plain') {
        document.getElementById('paste').textContent = new TextDecoder().decode(c);
        prettyPrint();
        setView('show_textpaste', 'header');
    } else if (isImage(mime)) {
        document.getElementById('displayimage').src = blobURL;
        setView('show_image', 'header');
    } else if (isVideo(mime)) {
        document.getElementById('displayvideo').src = blobURL;
        setView('show_video', 'header');
    } else {
        var fname = location.hash.split(/:/);
        if (fname[3]) {
            fname = fname[3];
        } else {
            fname = fname[0].substr(1)
        }
        document.getElementById('forceDownload').href = blobURL;
        document.getElementById('forceDownload').download = fname
        document.getElementById('fileext').textContent = fname
        setView('show_download', 'header');
    }
}

function readFile(f, cb) {
    var r = new FileReader();
    r.onload = function (e) {
        var contents = e.target.result;
        cb(contents, f.name);
    }
    r.readAsArrayBuffer(f);
}

function isImage(mime) {
    return ['image/jpg', 'image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/avif'].includes(mime);
}

function isVideo(mime) {
    return mime.startsWith("video/");
}

function pagestateFollowsHash() {
    var url = window.location + '';
    if (url.indexOf("#") != -1) {
        var d = url.split("#")[1].split(":");
        if (typeof d[0] !== 'undefined' && d[0].length > 1) {
            getBin(d[0], d[1], d[2]);
        }
    } else {
        setView('new_paste', 'upload');
    }
}

function setView() {
    var views = document.querySelectorAll('.view');
    [...views].forEach((v) => {
        v.style.display = 'none';
    });

    for (var view of arguments) {
        var el = document.querySelector('.view_' + view);
        el.style.display = 'block';
    }
}


function arrayBufferToBase64(buffer) {
    var binary = '';
    var bytes = new Uint8Array(buffer);
    var len = bytes.byteLength;
    for (var i = 0; i < len; i++) {
        binary += String.fromCharCode(bytes[i]);
    }
    return window.btoa(binary);
}



function base64ToArrayBuffer(base64) {
    var binary_string = window.atob(base64);
    var len = binary_string.length;
    var bytes = new Uint8Array(len);
    for (var i = 0; i < len; i++) {
        bytes[i] = binary_string.charCodeAt(i);
    }
    return bytes.buffer;
}


document.addEventListener("DOMContentLoaded", initPage);
window.onhashchange = pagestateFollowsHash;

function initPage() {

    window.addEventListener("paste", function (e) {
        if (e.clipboardData.items.length < 1) return;
        var file = e.clipboardData.items[0];
        if (file.kind != 'file') return;
        G_dataTransferFile = file.getAsFile();
        document.getElementById('uploadFrom').textContent = "Upload " + G_dataTransferFile.name
        setView('new_file_transfer', 'upload');
    });
    window.addEventListener("dragover", function (e) {
        e = e || event;
        e.preventDefault();
    }, false);
    window.addEventListener('drop', function (e) {
        console.log(e);
        e.stopPropagation();
        e.preventDefault();
        var files = e.dataTransfer.items; // Array of all files

        if (files < 1) return;
        var file = files[0];
        if (file.kind != 'file') return;
        G_dataTransferFile = file.getAsFile();
        document.getElementById('uploadFrom').textContent = "Upload " + G_dataTransferFile.name
        setView('new_file_transfer', 'upload');
    });

    pagestateFollowsHash();
}
