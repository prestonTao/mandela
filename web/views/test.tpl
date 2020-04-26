<!DOCTYPE html>
<html>
<head>
<!-- Standard Meta -->
<meta charset="utf-8" />
<meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1" />
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0">
<title>网络</title>
<link rel="icon" href="/static/img/favicon.png" type="image/x-icon" />


<script type="text/javascript" src="/static/js/jquery.min.js"></script>
<link rel="stylesheet" href="/static/dist/semantic.min.css">
<script src="/static/dist/semantic.min.js"></script>

<style type="text/css">
body {
    background-color: #FFFFFF;
}
.ui.menu .item img.logo {
    margin-right: 1.5em;
}
.main.container {
}
.wireframe {
    margin-top: 2em;
}
.ui.footer.segment {
    margin: 5em 0em 0em;
    padding: 5em 0em;
}
.content_img{
    max-width:100%;
    /*margin-top: 2em;*/
}
.border{
    border:red solid 1px;
}

.area{
    height:20px;
    font-weight: bold;
    font-size: 18px;
    color: #09c;
}
a, a:visited{
    text-decoration: none;
    color: #09c;
}
/*.hide{
    display:none;
}*/

/*@media only screen and (min-width: 180px) and (max-width: 800px) {
    .nav_max {
        display:block;
    }
    .nav_min {
        display:none;
    }
}*/

/*@media only screen and (min-width: 1000px) {
    .nav_max {
        display:block;
    }
    .nav_min {
        display:none;
    }
}*/
</style>

<script type="text/javascript">
$(function(){
    

$("#send").click(function(){
    // alert(1);
    // return;
    var id = $("#id").val();
    var content = $("#content").val();
    $.ajax({
        url: "/self/sendtextmsg",
        // cache: false,
        type:"POST",
        data: {"id":id, "content": content},
        dataType:"json",
        success: function(data){
            if (data.Code == 0){
            }else{
                alert("错误");
            }
        }
    });
});

$("#apply").click(function(){
    // alert(1);
    // return;
    var name = $("#name").val();
    $.ajax({
        url: "/self/applyname",
        // cache: false,
        type:"POST",
        data: {"name": name},
        dataType:"json",
        success: function(data){
            if (data.Code == 0){
            }else{
                alert("错误");
            }
        }
    });
});


$("#sendname").click(function(){
    // alert(1);
    // return;
    var name = $("#dstname").val();
    var content = $("#namecontent").val();

    $.ajax({
        url: "/self/sendmsgtoname",
        // cache: false,
        type:"POST",
        data: {"name": name, "content":content},
        dataType:"json",
        success: function(data){
            if (data.Code == 0){
            }else{
                alert("错误");
            }
        }
    });
});


$("#bt_test").click(function(){

    $.ajax({
        url: "/self/bttest",
        // cache: false,
        type:"POST",
        data: {},
        dataType:"json",
        success: function(data){
            if (data.Code == 0){
            }else{
                alert("错误");
            }
        }
    });
});


$("#regaccbt").click(function(){

    $.ajax({
        url: "/rpc",
        // cache: false,
        type:"POST",
        headers : {'user':'test','password':'testp','Content-Type':'application/json'},
        data: {},
        dataType:"json",
        success: function(data){
            if (data.Code == 0){
                alert("调用成功"+data);
            }else{
                alert("错误");
            }
        }
    });
});

});
</script>
</head>
<body>

    <div class="ui main container">
        <div>
            <a href="/static/index.html">云存储</a>
            <a href="/sharebox/page">共享盒子</a>
        </div>
        <!-- <div>你以为朝鲜的互联网是我们的过去，其实是我们的未来。</div> -->

        <div>本节点是否是超级节点 <%$.IsSuper%> &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;本机IP：<%$.Ip%></div>
        <div>根节点是否在线 <%$.RootExist%></div>

        
        <div>本节点id为 <%$.ID%></div>
        
        <div>超级节点id为 <%$.SuperId%></div>

        <div>伙伴ids</div>
        
        <%range $i, $e := $.ids%>
        <div><%$e%></div>
        <%end%>

        <div>本机保存的域名</div>
        <%range $i, $e := $.names%>
        <div><%$e.Name%></div>
            <%range $j, $ele := $e.Ids%>
            <div><%$ele.SuperPeerId%></div>
            <%end%>
        <%end%>

        <div>发送消息</div>
        <div>
            <input id="id" type="text" name="" style="width:500px;">消息内容<input id="content" type="text" name=""><input id="send" type="button" name="" value="发送">
        </div>
        <div>申请域名</div>
        <div>
            <input id="name" type="text" name="" style="width:500px;"><input id="apply" type="button" name="" value="申请">
        </div>


        <div>给域名发送消息</div>
        <div>
            <input id="dstname" type="text" name="" style="width:500px;">消息内容<input id="namecontent" type="text" name=""><input id="sendname" type="button" name="" value="发送">
        </div>
        <div>测试按钮</div>
        <div><input id="bt_test" type="button" name="" value="点击测试"></div>
        

        <div>注册域名</div>
        <div>
            <input id="regacccontent" type="text" name="" value="{&quot;method&quot;:&quot;namesin&quot;,&quot;params&quot;:{&quot;class&quot;:true,&quot;address&quot;:&quot;12FP28gjhN9cTmXxL4yd28VBMykQ3R&quot;,&quot;amount&quot;:1,&quot;gas&quot;:1,&quot;pwd&quot;:&quot;123456&quot;,&quot;name&quot;:&quot;test&quot;,&quot;netids&quot;:[&quot;12FP28gjhN9cTmXxL4yd28VBMykQ3R&quot;]}}" style="width:500px;"><input id="regaccbt" type="button" name="" value="注册">
        </div>

        
    </div>

</body>

</html>
