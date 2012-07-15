function getRandomPassword(){
    return CryptoJS.lib.WordArray.random(24).toString(CryptoJS.enc.Base64).replace("+","_");
}

function getContentEncrypted(password){
    return CryptoJS.AES.encrypt($('#r').val(),password + '' );
}

function decryptContent(c,pw) {
    return CryptoJS.AES.decrypt(c,pw);
}

function pasteIt(){
    var pass = getRandomPassword();
    var enccontent = getContentEncrypted(pass);
    $.post('/store', {r:enccontent+''}, function(data) {
            var res = data.split(/:/);
            if(res[0]=="OK"){
                $('#r').val("");
                window.location = findBaseURL() + "#" + res[1] + ":" + pass;
                getBin(res[1],pass);
            }
        });
}

function findBaseURL(){
    return (window.location+'').split(/#/)[0];
}

function getBin(id,password) {
    $.get("/get/" + id, function(data){
        data = decryptContent(data,password).toString(CryptoJS.enc.Utf8);
        showBin(data); 
    });
}

function showBin(c) {
    $("#paster").hide();
    $("#paste").text(c);
    prettyPrint();
    $("code,pre").show();
}

$(document).ready(function(){
    var url=window.location+'';
    if(url.indexOf("#") != -1){
        var d = url.split("#")[1].split(":");
        if(d[0]>=0){
            getBin(d[0],d[1]);
        }
    }else{
        $("#paster").show();
    }
});
