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
<link rel="stylesheet" href="/static/dist/semantic.min.css">
<script src="/static/dist/semantic.min.js"></script>

<script src="https://vuejs.org/js/vue.js"></script>

<style type="text/css">
body {
    background-color: #FFFFFF;
}
.ui.menu .item img.logo {
    margin-right: 1.5em;
}
.main.container {
    margin-top: 7em;
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


</head>
<body>
    
    <%template "nav.tpl" .%>

    

    <div class="ui main container">
        <div id="app">
        </div>

        
        <div>
            <a href="/store">云存储首页</a><br>
            <a href="/self/test">测试网络</a>
        </div>
    </div>
    
<script type="text/javascript">



var app = new Vue({
  el: '#app',
  data: {
    message: 'Hello Vue!'
  }
})

</script>
</body>

</html>
