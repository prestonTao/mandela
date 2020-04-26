<!DOCTYPE html>
<html>
<head>
<!-- Standard Meta -->
<meta charset="utf-8" />
<meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1" />
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0">
<title>云存储</title>
<link rel="icon" href="/static/img/Spy_24px.png" type="image/x-icon" />


<script type="text/javascript" src="/static/js/jquery.min.js"></script>
<link rel="stylesheet" href="/static/dist/semantic.min.css">
<script src="/static/dist/semantic.min.js"></script>

<link href="/static/js/jQuery-File-Upload-master/css/jquery.fileupload.css" rel="stylesheet" />
<link href="/static/js/jQuery-File-Upload-master/css/jquery.fileupload-ui.css" rel="stylesheet" />
<script src="/static/js/jQuery-File-Upload-master/js/vendor/jquery.ui.widget.js"></script>
<script src="/static/js/jQuery-File-Upload-master/js/jquery.fileupload.js"></script>
<script src="/static/js/jQuery-File-Upload-master/js/jquery.iframe-transport.js"></script>




<!-- <script src="https://vuejs.org/js/vue.js"></script> -->

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


/*#fileupload{
    width: 100px;
    border:red solid 1px;
}*/

.selectfile{
    height: 100px;
    padding-top: 30px;
    border: 1px dashed #ccc;
}

#uploadform{
    /*position: relative;*/
    width: 100px;
    margin-top: 50px;
    margin:0 auto;
}
.se2{
    /*width:100px;*/
    height:36px;
    position:absolute;
    /*top:338px;*/
    /*left:942px;*/
    z-index: 1;
    opacity: 0;
}
.se1{
    /*width:100px;*/
    height:36px;
    font-size:16px;
    color:#fff;
    background: #82939c;
    border-radius:5px;
    position:absolute;
    /*top:338px;*/
    /*left:942px;*/
}
.se1:hover{
    cursor: pointer;
}

.warn{
    align-content: center;
    text-align: center;
    border: 1px dashed #ccc;
}
</style>



<script type="text/javascript">
$(function(){

$('#fileupload').fileupload({
    url: "/store/addfile",
    done: function (e, data) {
        window.location.reload()
    }
});


});


</script>

</head>
<body>
    
    <%template "nav.tpl" .%>

    

    <div class="ui main container">
        <%if .CheckKey%>
        <%else%>
        <div class="warn">
            <a href="">没有钱包地址，点击生成</a>
        </div>
        <%end%>




        <div class="selectfile">
            <div id="uploadform" class="">
                <form  method='post' enctype="multipart/form-data">
                    <input class="se2" id="fileupload" type="file" name="files[]"/>
                    <label for="fileupload">
                        <input class="se1" type="button" value="选择文件" />
                    </label>
                </form>
            </div>
        </div>

        <div style="text-transform : uppercase; margin-top:60px;font-weight: 700;font-size: 18px;">All Local Files</div>
        <div style="border: 1px solid #ccc;padding: 20px;margin-top: 10px;">
            <div>
                <div class="" style="width:300px;">ID</div>
            </div>

            <div style="border: 1px solid #ccc;"></div>

            <%range $i, $e := $.Names%>
                <div>
                    <div class="" style="width:300px;"><%$e%></div>
                </div>
            <%end%>

            
        </div>
    </div>
    

</body>


</html>
