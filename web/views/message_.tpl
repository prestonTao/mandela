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

@media screen and (max-width:768px){
    .main{
        width: 100%;
        display: flex;
    }
}
@media screen and (min-width:769px) and (max-width:960px){
    .main{
        width: 768px;
        margin:0 auto;
        display: flex;
    }
}
@media screen and (min-width:961px){
    .main{
        width: 960px;
        margin:0 auto;
        display: flex;
        margin-top:2.7em;
    }
}
/*.main{
    margin-top:2.7em;
}*/
.menu{
    background-color: indigo;
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
.icon{
    flex:0 0 8em;
    height: 8em;
    background:url(/static/img/Spy_128px.png) no-repeat #fff;
    background-position: left 0em top 0em;
    position: relative;
}
.content{
    flex:10 1 auto;
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


/*body {
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

.ui.message{
    margin: 0;
}
.ui.message:last-child{
    margin-bottom: 1em;
}*/

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

window.friends = {
    info : {},
};


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
                // alert(name);
                // var info = window.friends.info.
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


var friends_list = new Vue({
    el: '#friends_list',
    data: {
        demo: [{Id: "山鹰登山社5", Index: 5, Content: "haha", Logs:[{Id: "123453224", Date: "2017-8-8 5:45", Content: "你好啊"}, {Id: "123453224", Date: "2017-8-8 5:45", Content: "我不好"}]},{Id: "山鹰登山社2", Index: 2, Content: "haha"}]
    },
    computed: {
        datas() {
            return this.demo.sort((a,b)=>a.Index-b.Index);
        }
    },
    methods: {
        showModel: function (index) {
            // alert("#item"+index+' .modal');
            // $('.modal').dimmer('show');
            // $('.segment').dimmer('show');

            $("#model"+index).modal('show');
        },
        buildIndex: function(index){
            return "item"+index;
        },
        buildModelIndex: function(index){
            return "model"+index;
        }
    }
});




getmsg();
    


$("#icbn_bt_search").click(function(){
    var name = $("#autocomplete").val();
    $.ajax({
        url: "/encyclopedia/search",
        // cache: false,
        type:"POST",
        data: {"Name":name},
        dataType:"json",
        success: function(data){
            if (data.Code == 0){
                window.location.href='/encyclopedia/plant/'+data.Id;
            }else if(data.Code == 2){
            }else if(data.Code == 1){
            }else{
            }
        }
    });
});



    

    $('.ui.search').search({
        apiSettings: {
            url: '/icbn/searchplant?term={query}'
        },
        // fields: {
        //     results : 'items',
        //     title   : 'name',
        //     url     : 'html_url'
        // },
        minCharacters : 1
    });


// var msg_logs = new Vue({
//     el: '.friends_list',
//     data: {
//         demo: [{Id: "山鹰登山社5", Index: 5, Content: "haha"},{Id: "山鹰登山社2", Index: 2, Content: "haha"}]
//     },
//     computed: {
//         datas() {
//             return this.demo.sort((a,b)=>a.Index-b.Index);
//         }
//     },
//     methods: {
//         showModel: function (index) {
//             // alert("#item"+index+' .modal');
//             // $('.modal').dimmer('show');
//             // $('.segment').dimmer('show');

//             $("#item"+index+' .ui.modal').modal('show');
//         },
//         buildIndex: function(index){
//             return "item"+index;
//         }
//     }
// });



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
</script>
</head>
<body class="blurring segment">
    

    <div class="menu">
        <a href="#" class="listitem">首页</a>
        <a href="#" class="listitem">论坛</a>
        <a href="#" class="listitem">博客</a>
        <!-- <a href="#" class="lastitem">帮助</a> -->
    </div>
    <div class="main">
        <div class="icon">
            <!-- <div class="tag">3</div> -->
        </div>
        <div class="content">
            <p>萌萌狗有各种形状和大小。有的小狗因为呆萌的表情惹人疼爱，有的则因为五短身材令人怜惜。甚至还有一些会因为巨大的体型也会显得傻缺。</p>
            <p>Many people also have their own barometers for what makes a cute dog.</p>
        </div>
    </div>


    

    <div class="ui main container">



        <div id="friends_list" class="ui divided items">


            <div v-for="(val, i) in datas" v-bind:id="buildIndex(i)" class="item">
                <a class="ui tiny image">
                    <img v-on:click="showModel(i)" src="/static/img/Spy_128px.png">
                </a>
                <div class="content">
                    <a class="header">{{val.Id}}</a>
                    <div class="description">
                        <p>萌萌狗有各种形状和大小。有的小狗因为呆萌的表情惹人疼爱，有的则因为五短身材令人怜惜。甚至还有一些会因为巨大的体型也会显得傻缺。</p>
                        <p>Many people also have their own barometers for what makes a cute dog.</p>
                    </div>
                </div>

                <!-- <div class="ui modal">
                    <i class="close icon"></i>
                    <div class="header">Profile Picture</div>
                    <div class="ui medium image">
                        
                        <div class="description">
                            <div class="ui divided items">

                                <div class="item" v-for="(one, j) in val.Logs">
                                    <div class="ui tiny image">
                                        <img src="">
                                    </div>
                                    <div class="middle aligned content">{{one.Content}}</div>
                                </div>
                            
                        </div>
                    </div>
                    <div class="actions">
                        <div class="ui black deny button">Nope</div>
                        <div class="ui positive right labeled icon button">Yep, that's me
                            <i class="checkmark icon"></i>
                        </div>
                    </div>
                </div> -->

                <div v-bind:id="buildModelIndex(i)" class="ui modal">
                    <i class="close icon"></i>
                    <div class="header">Profile Picture</div>
                    <div class="image content">
                        <div class="description">
                            <div v-for="(one, j) in val.Logs" class="" style="text-align: right;">
                                <div class="">{{one.Id}}</div>
                                <div class="ui compact message">
                                    <p>{{one.Content}}</p>
                                </div>
                            </div>

                            <!-- <div class="ui items">

                                <div class="item" v-for="(one, j) in val.Logs">
                                    <div class="">{{one.Id}}</div>
                                    <div class="ui compact message">
                                        <p>{{one.Content}}</p>
                                    </div>
                                </div>
                            </div> -->
                            
                        </div>
                    </div>
                    <div class="actions">
                        <div class="ui black deny button">Nope</div>
                        <div class="ui positive right labeled icon button">Yep, that's me
                            <i class="checkmark icon"></i>
                        </div>
                    </div>
                </div>
                
            </div>


<!-- <div class="description">
                            <div class="ui divided items">

                                <div class="item" v-for="(one, j) in val.Logs">
                                    <div class="ui tiny image">
                                        <img src="">
                                    </div>
                                    <div class="middle aligned content">{{one.Content}}</div>
                                </div>
                            
                        </div> -->
            

        </div>



    </div>






</body>

</html>
