function getnewaddress(){
    $.ajax({
        url: '/v1/object/',
        type: 'POST',
        beforeSend: function (request) {
          request.setRequestHeader("user", "test");
          request.setRequestHeader("password", "testp");
        },
        data: {
          method: "getinfo",
          params: { "id": 10 }
        },
        dataType: 'json',
        success: function (res) {
          var time = new Date(); 
          var accountstr = ` <li class="am-comment">
                      <a href="#">
                        <img src="face/1.jpeg" alt="" class="am-comment-avatar" width="48" height="48">
                      </a>
                      <div>
                        <header class="am-comment-hd">
                          <div class="am-comment-meta">
                            <a href="#" class="am-comment-author">账户1 </a>
                            <time>   `+ time.toString().substring(10, 25) + `</time>
                          </div>
                        </header>
                        <div class="am-comment-bd">
                          <div class="am-g am-margin-top">
                            <div class="am-u-sm-4 am-u-md-1 am-text-right">可使用</div>
                            <div class="am-u-sm-8 am-u-md-11">
                              <label>`+ res.result.balance + ` </label>
                            </div>
                          </div>
    
                          <div class="am-g am-margin-top">
                            <div class="am-u-sm-4 am-u-md-1 am-text-right">
                              等待中
                            </div>
                            <div class="am-u-sm-8 am-u-md-11">
                              <label> 0.25821232 </label>
                            </div>
                          </div>
    
                          <div class="am-g am-margin-top">
                            <div class="am-u-sm-4 am-u-md-1 am-text-right">
                              总额
                            </div>
                            <div class="am-u-sm-8 am-u-md-11">
                              <label>`+ res.result.blocks + ` </label>
                            </div>
                          </div>
                        </div>
                      </div>
                    </li>`;
          $("#accountDiv").append(accountstr);
        }
      })

}