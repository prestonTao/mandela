var centerUrl = "http://39.104.112.203:8080";
function checkPasswords() {
    var pass1 = document.getElementById("password");
    var repassword = document.getElementById("repassword");
    var repasswordWrap = document.getElementById("repasswordWrap");
    var repasswordTip = document.getElementById("repasswordTip");
    if (repasswordTip != null) {
        repasswordWrap.removeChild(repasswordTip);//
    }
    if (pass1.value == "") {
        repasswordTip = document.createElement("small");
        repasswordTip.id = "repasswordTip"
        repasswordTip.innerText = "密码不能为空";
        repasswordWrap.appendChild(repasswordTip);
        return false

    }
    if (pass1.value != repassword.value) {
        repasswordTip = document.createElement("small");
        repasswordTip.id = "repasswordTip"
        repasswordTip.innerText = "两次输入的密码不一致";
        repasswordWrap.appendChild(repasswordTip);
        return false
    }
    return true
}
//设置cookie
function setCookie(name, value, day) {
    var date = new Date();
    date.setDate(date.getDate() + day);
    document.cookie = name + '=' + value + ';expires=' + date;
};
//获取cookie
function getCookie(name) {
    var reg = RegExp(name + '=([^;]+)');
    var arr = document.cookie.match(reg);
    if (arr) {
        return arr[1];
    } else {
        return '';
    }
};
//删除cookie
function delCookie(name) {
    setCookie(name, "", -1);
};
function getCode() {
    var sessionid = getCookie("Sessionid");

    $.ajax({
        url: centerUrl + '/api/user/sendSMS/' + $("#telInput").val(),
        type: 'get',
        headers: {
            Sessionid: sessionid
        },
        dataType: 'json',
        success: function (res) {
            if (res.status != 200) {
                alert(res.msg);
            } else {
                leveltime = 60;
                currentDaoJIshi = setInterval("setDaoJiShi()", 1000)
                console.log(res);
                setCookie("smscode", res.data, 0.1);
            }

        }
    })
}
var currentDaoJIshi;
var leveltime = 60;
function setDaoJiShi() {

    console.log(document.getElementById('huoquyanzhengma').value);
    leveltime--;
    if (leveltime == 0) {
        window.clearInterval(currentDaoJIshi);
        document.getElementById("huoquyanzhengma").innerHTML = "发送验证码";
      
    } else {
        document.getElementById("huoquyanzhengma").innerHTML ="等待" + leveltime + "后重发";  
    }
}
function submitWebform() {

    if (document.getElementById('readed').checked == true) {

        submitform("index.html");
    } else {
        alert("请先阅读用户协议");
    }

}

var myObj = {};
function submitform(redictUrl) {
    if (checkPasswords() == false) {

        return
    }
    var currentcode = getCookie("smscode");
    var codeinput = $("#codeinput").val();
    if (currentcode != codeinput) {
        alert("验证码错误");
        return
    }
    var user = {};
    user.code = parseInt(currentcode);
    user.username = $("#usernameInput").val();
    user.password = hex_md5($("#password").val());
    user.name = $("#name").val();
    user.email = $("#email").val();
    user.tel = $("#telInput").val();
    user.firstname = $("#firstname").val();
    user.lastname = $("#lastname").val();
    user.intro = $("#intro").val();
    var sessionid = getCookie("Sessionid");
    $.ajax({
        url: centerUrl + '/api/user/register',
        type: 'POST',
        headers: {
            Sessionid: sessionid
        },
        data: JSON.stringify(user),
        contentType: "application/json",
        dataType: "json",
        success: function (res) {
            if (res.status != 200) {
                alert(res.msg);
                return;
            } else {
                setCookie('Sessionid', res.msg, 1);
                setCookie('currentuser', JSON.stringify(res.data), 1);
                $(window.location).attr('href', redictUrl);
            }

        }
    })
}