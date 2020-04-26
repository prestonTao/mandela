<!DOCTYPE html>
<html>
<head>
<!-- Standard Meta -->
<meta charset="utf-8" />
<meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1" />
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0">
<title>网络</title>
<link rel="icon" href="/static/img/Spy_24px.png" type="image/x-icon" />


<script type="text/javascript" src="/static/js/jquery.min.js"></script>
<!-- <link rel="stylesheet" href="/static/dist/semantic.min.css">
<script src="/static/dist/semantic.min.js"></script>

<script src="https://vuejs.org/js/vue.js"></script> -->

<style type="text/css">
html,body {
    margin:0;
    padding:0;
    background-color: #FFFFFF;
    font-family: Lato,'Helvetica Neue',Arial,Helvetica,sans-serif;
    /*font: 150%/1.6 Baskerville, Palatino, serif;*/
}

@media screen and (max-width:768px){
    .content{
        width: 100%;
        padding-top: 1em;
    }
}
@media screen and (min-width:769px) and (max-width:960px){
    .content{
        width: 768px;
        margin:0 auto;
        padding-top: 1em;
    }
}
@media screen and (min-width:961px){
    .content{
        width: 960px;
        margin:0 auto;
        padding-top: 1em;
    }
}


/*.main{
    margin-top:2.7em;
}*/
.menu{
    background-color: #1B1C1D;
    font-size: 1em;
    min-height: 2.7em;
    display: flex;
    /*让item反序排列*/
    /*flex-direction: row-reverse;*/
    align-items: center;
    padding: 0 1em;
    position: fixed;
    width:100%;
    z-index: 2;
    top:0;
}
.seat{
    margin-top: 2.7em;
}
.listitem, .lastitem{
    color: #ebebeb;
    text-decoration: none;
}
.listitem{
    margin-right: 1em;
}
.lastitem{
    margin-left: auto;
}

.border{
    border: red solid 1px;
}
/*main{
    width: 100%;
    margin:0 auto;
    display: flex;
}*/

.items{
    display: flex;
    flex-direction: column-reverse;
}
.item{
    display: flex;
}
.icon{
    flex:0 0 6em;
    height: 6em;
    background:url(/static/img/Spy_96px.png) no-repeat #fff;
    /*background-position: left 0em top 0em;*/
    position: relative;
}
.content{
    flex:10 1 auto;
    padding-left: 0.8em;
}
.content a{
    font-weight: 700;
}

.tag{
    width: 1.1em;
    height: 1.1em;
    /*padding: 1em;*/
    border-radius:10em;
    background-color: #f43531;
    position: absolute;
    right: 0;
    text-align: center;
    line-height: 1.2em;
    color: #fff;
    font-family: PingFang SC,Hiragino Sans GB,Arial,Microsoft YaHei,Helvetica;
    /*font-size: 0.875em;*/
}

.dialog {
    position: fixed;
    top: 50%; left: 50%;
    z-index: 1;
    width: 10em;
    padding: 2em;
    margin: -5em;
    border: 1px solid silver;
    border-radius: .5em;
    box-shadow: 0 .2em .5em rgba(0,0,0,.5),
                0 0 0 100vmax rgba(0,0,0,.2);
    background: white;
}
.dialog:not([open]) {
    display: none;
}

.main {
    transition: .6s;
    background: white;
}
.main.de-emphasized {
    /*-webkit-filter: blur(3px);*/
    filter: blur(3px);
}
.dialog_msg{
    position: fixed;
    top: 0%; left: calc((100% - 31em) / 2);
    z-index: 1;
    width: 30em;
    padding: 0.5em;
    margin: 0em;
    border: 1px solid silver;
    border-radius: .5em;
    box-shadow: 0 .2em .5em rgba(0,0,0,.5),
                0 0 0 110vmax rgba(0,0,0,.2);
    background: white;
}
.dialog_msg:not([open]) {
    display: none;
}


/*a, a:visited{
    text-decoration: none;
    color: #09c;
}*/

.msgcontent{
    display: flex;
    flex-direction: column;
}
.msgrecv{
    margin-top: .2em;
    padding: .5em;
    background: #f8f8f9;
    font-size: .8em;
    border-radius: .28571429em;
    box-shadow: 0 0 0 1px rgba(34,36,38,.22) inset, 0 0 0 0 transparent;
    display: inline-block;
}
.msgsend{
    margin-top: .2em;
    padding: .5em;
    background: #CCE2FF;
    font-size: .8em;
    border-radius: .28571429em;
    box-shadow: 0 0 0 1px #A9D5DE inset, 0 0 0 0 transparent;
    color: #276F86;
    display: inline-block;
}
.msgfloor{
    margin-top: .5em;
    text-align: right;
}


.newmessage{
    display:none;
}

</style>

<script type="text/javascript">
$(function(){


$("#bt_add_friend").click(function(){
    $("#dialog_addfriend").attr('open', '');
    $(".main").addClass('de-emphasized');
});

window.friends = {};
<% range $i, $elem := $.Fs %>
window.friends[<% $elem.Id %>] = <% $elem.Id %>;
<% end %>



function getmsg(){
    $.ajax({
        url: "/self/msg/getmsg",
        // cache: false,
        type:"POST",
        // data: {"Name":name},
        dataType:"json",
        success: function(data){
            // alert(data.Code);
            if (data.Code == 0){
                // window.location.href='/encyclopedia/plant/'+data.Id;
                // data.Content
                // var name = data.Name;
                // alert(data.Content);
                // var info = window.friends.info.
                
                if(typeof(window.friends[data.Id])=="undefined"){
                    $(".newmessage").css("display","block");
                    // alert("有新消息");
                }else{
                    // alert("显示新消息");
                    $("#tag_"+data.Id).css("display","block");
                    var n = $("#tag_"+data.Id).text();
                    $("#tag_"+data.Id).text(parseInt(n) + 1);
                    $("#item_"+data.Id).css("order",data.Index);
                }
            }else if(data.Code == 2){
            }else if(data.Code == 1){
            }else{
            }
            getmsg();



            // friends_list.demo[0].Index = 2;


            // var newObj = friends_list.demo[1];
            // newObj.Index = 1
            // Vue.set(friends_list.demo, 1, newObj);

            

        }
    });
}





getmsg();
    


$("#bt_add").click(function(){
    var id = $("#input_id").val();
    $.ajax({
        url: "/self/friend/add",
        // cache: false,
        type:"POST",
        data: {"ID":id},
        dataType:"json",
        success: function(data){
            if (data.Code == 0){
                $("#dialog_addfriend").removeAttr('open');
                $(".main").removeClass('de-emphasized');
                $("#input_id").val("");
                window.location.reload();

            }else if(data.Code == 2){
            }else if(data.Code == 1){
            }else{
            }
        }
    });
});


$("#bt_cancel").click(function(){
    $("#dialog_addfriend").removeAttr('open');
    $(".main").removeClass('de-emphasized');
    $("#input_id").val("");
});



    





});
function quit_bt(){
    $.ajax({
        url: "/logout",
        success: function(data){
            window.location.reload();
        }
    });
    return false;
}
function openMsgDialog(index){
    // alert(index);
    
    $("#dialog_msg_"+index).attr('open', '');
    $(".main").addClass('de-emphasized');
}

function closeMsgDialog(index){
    $("#dialog_msg_"+index).removeAttr('open');
    $(".main").removeClass('de-emphasized');
}
</script>
</head>
<body>
    
    <div class="main">
        <div class="menu">
            <a href="/" class="listitem">首页</a>
            <a href="/self/msg" class="listitem">消息</a>
            <!-- <a href="#" class="listitem">博客</a> -->
            <!-- <a href="#" class="lastitem">帮助</a> -->
        </div>

        <div class="content">
            <div class="seat"></div>
            <div class=""><input id="bt_add_friend" type="button" name="" value="添加好友"></div>

            <div class="newmessage"><button class="positive ui button" onclick="window.location.reload()">有新消息</button></div>

            <div class="items">
                <% range $i, $elem := $.Fs %>
                <div id="item_<%$elem.Id%>" style="order:<%$elem.Order%>;" class="item">
                    <div class="icon">
                        <div id="tag_<%$elem.Id%>" class="tag" <% if eq $elem.Unread 0 %>style="display: none;"<% end %>><% $elem.Unread %></div>
                    </div>
                    <div class="content">
                        <a class="header"><% $elem.Id %></a>
                        <p>萌萌狗有各种形状和大小。有的小狗因为呆萌的表情惹人疼爱，有的则因为五短身材令人怜惜。甚至还有一些会因为巨大的体型也会显得傻缺。</p>
                        <p>Many people also have their own barometers for what makes a cute dog.</p>
                        <input type="button" name="" onclick="openMsgDialog(<%$i%>)" value="发送消息">
                    </div>
                </div>
                <% end %>
            </div>
        </div>
    </div>

    <div id="dialog_addfriend" class="dialog">
        <div>
            添加ID:<input id="input_id" type="text" name="">
        </div>
        <div>
            <input id="bt_cancel" type="button" name="" value="取消"><input id="bt_add" type="button" name="" value="确认">
        </div>
    </div>

    <% range $i, $elem := $.Fs %>
    <div id="dialog_msg_<% $i %>" class="dialog_msg">
        <div class="msgcontent">
            <div class="">
                <div class="msgrecv">你好啊</div>
            </div>
            <div class="">
                <div class="msgsend">你好</div>
            </div>
        </div>
        <div class="msgfloor">
            <input type="button" name="" onclick="closeMsgDialog(<%$i%>)" value="取消"><input id="bt_add" type="button" name="" value="确认">
        </div>
    </div>
    <% end %>




    





</body>

</html>
