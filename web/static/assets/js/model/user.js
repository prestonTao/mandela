var centerUrl = "http://39.104.112.203:8080";
window.onload = function () {
  var oUser = document.getElementById('usernameInput');
  var oPswd = document.getElementById('passwordInput');
  var oRemember = document.getElementById('remember');
  if (oUser != null ){
    if ( getCookie('username')){
      oUser.value = getCookie('username');
    }
  }

  if (oPswd != null ){
    if ( getCookie('password')){
      oPswd.value = getCookie('password');
    }
  }

  //页面初始化时，如果帐号密码cookie存在则填充 
  if ( oPswd != null &&   getCookie('password')) { 
    oRemember.checked = true;
  }
 
  if (oUser != null) { 
    //复选框勾选状态发生改变时，如果未勾选则清除cookie
    oRemember.onchange = function () { 
      if (!this.checked) {
        delCookie('username');
        delCookie('password');
       oPswd.value = "";
      }
    };
  }  
  //表单提交事件触发时，如果复选框是勾选状态则保存cookie

};
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
//----------------------
var cryptoPassword ="";
function pswdOnFocus() {
  delCookie('password');
  cryptoPassword = "";
  $("#passwordInput").val("");
  console.log("delete cookie:", getCookie('password'));
}

function loginFunc(redictUrl) {
  if (redictUrl == null) {
    redictUrl = './index.html'
  }
  code = $("#codeinput").val();
  cryptoPassword= getCookie('password');
  username = $("#usernameInput").val();
  if (cryptoPassword == ""){ //&&$("#passwordInput").val()!=""
    cryptoPassword = hex_md5($("#passwordInput").val()); 
  }

  var res = verifyCode.validate(code);
  if (res) {
    // console.log("验证码正确");
  } else {
    alert("验证码错误");
    return
  }

  if (username == "" ) {
    alert("用户名必填");
    return
  }
  console.log("cryptoPassword:",cryptoPassword);
  if ( cryptoPassword == "") {
    alert("密码必填");
    return
  }
  $.ajax({
    url: centerUrl + '/api/user/login',
    type: 'get',
    data: {
      username: username,
      password: cryptoPassword,
      code: code,
    },
    dataType: 'json',
    success: function (res) {
      if (res.status != 200) {
        alert(res.msg);
        if (redictUrl == './index.html') {
          $(window.location).attr('href', './login.html');
        } else {
          $(window.location).attr('href', './web-login.html');
        }

      } else {
        setCookie('Sessionid', res.msg, 1);
        setCookie('currentuser', JSON.stringify(res.data), 1); 
        var oUser = document.getElementById('usernameInput');
        var oPswd = document.getElementById('passwordInput');
        var oRemember = document.getElementById('remember');
        if (oUser != null && oRemember != null && oRemember.checked) {
          setCookie('username', oUser.value, 7); //保存帐号到cookie，有效期7天
          setCookie('password', cryptoPassword, 7); //保存密码到cookie，有效期7天
        }
        $(window.location).attr('href', redictUrl);
      }
    }
  })
};
 
function getCode() {
  var sessionid = getCookie("Sessionid");
  var currentuser = getCookie('currentuser');
  var usernamevalue = "";
  if (currentuser == "") {
    usernamevalue = $("#usernameInput").val();
  } else {
    currentuser = jQuery.parseJSON(currentuser);
    usernamevalue = currentuser.tel;
  }
  $.ajax({
    url: centerUrl + '/api/user/sendSMS/' + usernamevalue,
    type: 'get',
    headers: {
      Sessionid: sessionid
    },
    dataType: 'json',
    success: function (res) {
      if (res.status != 200) {
        alert(res.msg);
      } else {
        console.log(res); 

      }

    }
  })
}
function myAddressBook() {

  $.ajax({
    url: centerUrl + '/api/address/mine',
    headers: {
      Sessionid: sessionid
    },
    type: 'GET',
    dataType: 'json',
    success: function (res) {
      if (res.status == 500) {
        return
      }
      if (res.data == null) {
        return
      }
      res.data.forEach(function (item, index, input) {
        displayAccount(item);
      })
    }
  })
};
function logoutFunc(redictUrl) {
  var r = confirm("真的退出?");
  if (r == true) {
    doLogoutFunc(redictUrl);
  }
};
function doLogoutFunc(redictUrl) {

  if (redictUrl == null) {
    redictUrl = './login.html'
  }
  var sessionid = getCookie("Sessionid")
  $.ajax({
    url: centerUrl + '/api/user/logout',
    type: 'GET',
    dataType: 'json',
    headers: {
      Sessionid: sessionid
    },
    success: function (res) {
      delCookie("Sessionid")
      delCookie("currentuser") 
      $(window.location).attr('href', redictUrl);
    }
  })
}


function clearUsernameInput() {
  document.getElementById("usernameInput").value = "";
}

//--------------------------------------- 资料 -----------------------------------------------
function openInfomationPanel() {
  var item = {};
  var sessionid = getCookie("Sessionid");
  $.ajax({
    url: centerUrl + '/api/user/whoami',
    type: 'GET',
    headers: {
      Sessionid: sessionid
    },
    dataType: 'json',
    success: function (res) {
      if (res.status == 500) {
        return
      } 
      if (res.data == null) {
        alert("没有登录");
        return
      }
      item = res.data;
      var panel = '  <div class="hw-overlay" id="openInfoPanel-layer" style="display:none;z-index:99999;">'+
      ' <div class="hw-layer-wrap" style=" width:450px;" >'+
      ' <h3 style="margin:-50px;border: 2px solid #1999e3;'+
      ' padding-left:20px; width:auto;'+
      ' color: white; '+
      ' font-family: MicrosoftYaHei;'+
      ' font-size: 24px;'+
      ' font-weight: normal;'+
      ' font-stretch: normal;'+
      ' height: 46px;'+
      ' background-color: #1999e3; ">我的资料 <label   style=" cursor:  pointer;color:#fff; margin-left:270px; " onclick="cancelPersonInfopanel()">╳</label> </h3 >'+
      ' <br>'+
      '  <br> '+
 
      '  <div class="row">'+

      '  <form class="am-form" method="POST"> '+
      '  <div class="am-g am-margin-top"> '+
      '  </div>'+
      ' <div class="am-g am-margin-top">'+
      '  <div class="am-u-sm-3 am-u-md-3  am-text-right">      登录名:'+
      '   </div>'+
      '   <div class="am-u-sm-9 am-u-md-9  ">'+
      '    <input id="password" type="text" readonly value=' + item.username + ' required class="am-input-sm am-u-sm-9 am-u-md-8 ">'+
      '   </div>'+
      '   </div>'+
 

      '  <div class="am-g am-margin-top">'+
      ' <div class="am-u-sm-3 am-u-md-3  am-text-right">   邮件:'+
      '  </div>'+
      '  <div class="am-u-sm-9 am-u-md-9  ">'+
      '    <input type="text" id="email" placeholder="你的电子邮件 / Email" value=' + item.email + '  required class="am-input-sm am-u-sm-9 am-u-md-8 ">'+
      '  </div>'+
      '   </div>'+
    
      ' <div class="am-g am-margin-top">'+
      '   <div class="am-u-sm-3 am-u-md-3 am-text-right">'+
      '     电话:'+
      '   </div>'+
      '   <div class="am-u-sm-9 am-u-md-9  ">'+
      '     <input type="text" id="tel" placeholder="电话 / Telephone" value=' + item.tel + '  required class="am-input-sm am-u-sm-9 am-u-md-8 ">'+
    '   </div>'+
    '  </div>'+
    '  <div class="am-g am-margin-top">'+
    '   <div class="am-u-sm-3 am-u-md-3  am-text-right">'+
    '    注册于:'+
    '   </div>'+
    '   <div class="am-u-sm-9 am-u-md-9  ">'+
    '  <input type="text" readonly  value=' + item.reg_time + '  required class="am-input-sm am-u-sm-9 am-u-md-8 "> '+
    '   </div>'+
    ' </div>'+
    ' <div class="am-g am-margin-top">'+
    '   <div class="am-u-sm-3 am-u-md-3  am-text-right">  简介:'+
    '  </div>'+
    '   <div class="am-u-sm-9 am-u-md-9  ">'+
    '     <textarea class="" rows="5" id="intro" placeholder="输入个人简介"  class="am-input-sm am-u-sm-9 am-u-md-8 ">' + item.comments + '</textarea>'+
    '     <small>250字以内写出你的梦想.</small>'+
    '   </div>'+
    '  </div>'+

 
    ' </form>'+
    '   <br>'+
    '  <button type="button" onclick="updateUserinfo()" class="am-btn am-fl am-btn-primary am-btn-xs  hwLayer-ok">确  定'+
    '   </button>'+
    '    <button type="reset" class="am-btn am-btn-primary am-btn-xs am-fr  hwLayer-cancel" onclick="cancelPersonInfopanel()">取 消</button>'+

    '   </div>'+
    '  </div>'+
    ' </div>'; 

      $("#mainBody").prepend(panel);
      $("#openInfoPanel-layer").show();
      var topnum = 10;
      $("#openInfoPanel-layer").find(".hw-layer-wrap").css("margin-top", topnum);
    }
  })

}

function updateUserinfo() {
  var user = {};
  user.username = $("#usernameInput").val();
  user.name = $("#name").val();
  user.email = $("#email").val();
  user.tel = $("#tel").val();
  user.firstname = $("#firstname").val();
  user.lastname = $("#lastname").val();
  user.comments = $("#intro").val();
  var sessionid = getCookie("Sessionid");
  $.ajax({
    url: centerUrl + '/api/user/update',
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
      } else { 
        cancelPersonInfopanel();
        alert("修改成功");
      }

    }
  })
}
function cancelPersonInfopanel() {
  $("#openInfoPanel-layer").hide();
  $("#openInfoPanel-layer").remove();
}
// -----------------------------------  设置  -----------------------------------------------

function openSettingPanel() {
  var panel = '   <div class="hw-overlay" id="setting-layer" style="display:none;z-index:99999;">'+
  '  <div class="hw-layer-wrap" style=" width:450px;" >'+
  '  <h3 style="margin:-50px;border: 2px solid #1999e3;'+
  '  padding-left:20px; width:auto;'+
  '  color: white; '+
  '  font-family: MicrosoftYaHei;'+
  '  font-size: 24px;'+
  '  font-weight: normal;'+
  '  font-stretch: normal;'+
  '  height: 46px;'+
  '  background-color: #1999e3; ">设置<label   style=" cursor:  pointer;color:#fff; margin-left:270px; " onclick="cancelPersonInfopanel()">╳</label> </h3 >'+
  '  <br>'+
  '   <br>  '+
  '    <div class="row">'+

  '    <form class="am-form" method="POST"> '+
  '    <div class="am-g am-margin-top"> '+
  '   </div>'+
  '    <div class="am-g am-margin-top">'+
  '     <div class="am-u-sm-5 am-u-md-5  am-text-right">  记住密码: '+
  '      </div>'+
  '      <div class="am-u-sm-5 am-u-md-5  ">'+
  '      <label><input name="remember" type="checkbox" checked value="1" />记住密码 </label>'+
  '      </div>'+
  '    </div>'+

  '    <div class="am-g am-margin-top">'+
  '     <div class="am-u-sm-5 am-u-md-5  am-text-right">  交易费用自动扣除:'+
  '      </div>'+
  '      <div class="am-u-sm-5 am-u-md-5  " id="repasswordWrap">'+
  '     <input name="sex" type="checkbox" checked value="1" />自动扣除</label>'+
  '     </div>'+
  '    </div> '+

    
 
  '   </form>'+
  '     <br>'+
  '      <button type="button" onclick="createAccount()" class="am-btn am-fl am-btn-primary am-btn-xs  hwLayer-ok">确 定'+
  '     </button>'+
  '     <button type="reset" class="am-btn am-btn-primary am-btn-xs am-fr  hwLayer-cancel" onclick="cancelSettingpanel()">取 消</button>'+

  '    </div>'+
  '  </div>'+
  '  </div>';

  $("#mainBody").prepend(panel);
  $("#setting-layer").show();
  var topnum = 10;
  $("#setting-layer").find(".hw-layer-wrap").css("margin-top", topnum);
}

function cancelSettingpanel() {
  $("#setting-layer").hide();
  $("#setting-layer").remove();
}

//----------------------修改密码-------------------------
function openUpdatePswdPanel() {
  var item = {};


  var panel = '   <div class="hw-overlay" id="openUpdatePswdPanel-layer" style="display:none;z-index:99999;">'+
  '  <div class="hw-layer-wrap" style=" width:450px;" >'+
  '  <h3 style="margin:-50px;border: 2px solid #1999e3;'+
  '  padding-left:20px; width:auto;'+
  '  color: white; '+
  '  font-family: MicrosoftYaHei;'+
  '  font-size: 24px;'+
  '  font-weight: normal;'+
  '  font-stretch: normal;'+
  '  height: 46px;'+
  '   background-color: #1999e3; ">修改密码 <label   style=" cursor:  pointer;color:#fff; margin-left:270px; " onclick="cancelUpdatepswdpanel()">╳</label> </h3 >'+
  '   <br>'+
  '   <br>  '+
  '    <div class="row"> '+
  '    <form  > '+
  '    <div class="am-g am-margin-top"> '+
   '    </div> '+
   '    </form>  '+
   '    <div class="am-g am-margin-top">'+
   '    <div class="am-u-sm-3 am-u-md-3  am-text-right">   验证码:'+
   '   </div>'+
   '    <div class="am-u-sm-9 am-u-md-9  ">'+
   '    <input type="text" id="updatePswdCodeinput" >'+
   '      <button onclick="getCode()" style="width: 110px;  '+
   '      height: 41px; text-align: center; '+
   '      background-color: #c6cdd1; font-size: 14px;'+
   '  font-weight: normal;'+
   '  font-stretch: normal;'+
   '  letter-spacing: 0px;'+
   '  color: #f5f6f7;'+
   '  border: solid 1px #c6cdd1;"> 获取验证码'+
   '  </button>'+

   '    </div>'+
   '  </div>'+
   '    <div class="am-g am-margin-top">'+
   '     <div class="am-u-sm-3 am-u-md-3  am-text-right">  旧密码:'+
   '      </div>'+
   '     <div class="am-u-sm-9 am-u-md-9  ">'+
   '       <input id="oldpasswordInput" type="text"  required class="am-input-sm am-u-sm-9 am-u-md-8 ">'+
   '     </div>'+
   '   </div>'+
  
   '   <div class="am-g am-margin-top">'+
   '   <div class="am-u-sm-3 am-u-md-3  am-text-right">   新密码:'+
   '  </div>'+
   '    <div class="am-u-sm-9 am-u-md-9  ">'+
   '     <input id="newpasswordInput" type="text" required class="am-input-sm am-u-sm-9 am-u-md-8 ">'+
   '   </div>'+
   '  </div>'+
 

   '    <br>'+
   '   <button type="button" onclick="doUpdatePswd()" class="am-btn am-fl am-btn-primary am-btn-xs  hwLayer-ok">确   定'+
   '     </button>'+
   '     <button type="reset" class="am-btn am-btn-primary am-btn-xs am-fr  hwLayer-cancel" onclick="cancelUpdatepswdpanel()">取 消</button>'+

   '   </div>'+
   '  </div>'+
   '  </div>'; 
  $("#mainBody").prepend(panel);
  $("#openUpdatePswdPanel-layer").show();
  var topnum = 10;
  $("#openUpdatePswdPanel-layer").find(".hw-layer-wrap").css("margin-top", topnum);


}
function doUpdatePswd() {
  var currentuser = jQuery.parseJSON(getCookie('currentuser'));
  if (currentuser == null || currentuser == "") {
    alert("非法用户");
    return;
  }
  if (currentuser.password != hex_md5($("#oldpasswordInput").val())) {
    alert("密码验证错误");
    return;
  }

  var sessionid = getCookie("Sessionid");
  $.ajax({
    url: centerUrl + '/api/user/updatePswd',
    type: 'GET',
    headers: {
      Sessionid: sessionid
    },
    data: {
      code: $("#updatePswdCodeinput").val(),
      tel: currentuser.tel,
      newpswd: hex_md5($("#newpasswordInput").val()),
    },
    dataType: 'json',
    success: function (res) {
      if (res.status == 500) {
        alert(res.msg);
        return
      } 
      if (res.data == null) {
        alert("没有登录");
        return
      }
      alert("修改成功");
      cancelUpdatepswdpanel();
    }
  })
}
function cancelUpdatepswdpanel() {
  $("#openUpdatePswdPanel-layer").hide();
  $("#openUpdatePswdPanel-layer").remove();
}
