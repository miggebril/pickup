var App = function () {

    var currentPage = ''; 
    var pageID = '';
    var collapsed = false;
    var is_mobile = false;
    var is_mini_menu = false;
    var is_fixed_header = false;
    var responsiveFunctions = [];
    var firstLoad = true;

    var base64_chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_";
    var encodeObjectID = function binary_to_base64(input) {
      var binary = new Array();
      for (var i=0; i<input.length/2; i++) {
        var h = input.substr(i*2, 2);
        binary[i] = parseInt(h,16);        
      }

      var ret = new Array();
      var i = 0;
      var j = 0;
      var char_array_3 = new Array(3);
      var char_array_4 = new Array(4);
      var in_len = binary.length;
      var pos = 0;
  
      while (in_len--)
      {
          char_array_3[i++] = binary[pos++];
          if (i == 3)
          {
              char_array_4[0] = (char_array_3[0] & 0xfc) >> 2;
              char_array_4[1] = ((char_array_3[0] & 0x03) << 4) + ((char_array_3[1] & 0xf0) >> 4);
              char_array_4[2] = ((char_array_3[1] & 0x0f) << 2) + ((char_array_3[2] & 0xc0) >> 6);
              char_array_4[3] = char_array_3[2] & 0x3f;
  
              for (i = 0; (i <4) ; i++)
                  ret += base64_chars.charAt(char_array_4[i]);
              i = 0;
          }
      }
  
      if (i)
      {
          for (j = i; j < 3; j++)
              char_array_3[j] = 0;
  
          char_array_4[0] = (char_array_3[0] & 0xfc) >> 2;
          char_array_4[1] = ((char_array_3[0] & 0x03) << 4) + ((char_array_3[1] & 0xf0) >> 4);
          char_array_4[2] = ((char_array_3[1] & 0x0f) << 2) + ((char_array_3[2] & 0xc0) >> 6);
          char_array_4[3] = char_array_3[2] & 0x3f;
  
          for (j = 0; (j < i + 1); j++)
              ret += base64_chars.charAt(char_array_4[j]);
  
          while ((i++ < 3))
              ret += '=';
  
      }
  
      return ret;
    };

    var handleChat = function() {
        var conn;
        var msg = $("#ChatForm input");
        var log = $("#ChatLog");
        var user = $("#ChatLog").data("uid");
        function appendLog(data) {
            var d = log[0]
            var doScroll = d.scrollTop == d.scrollHeight - d.clientHeight;
            var message = $("<div class='chat-message'><img class='message-avatar' src='/users/"+user+"/profile' alt='' ><div class='message'><a class='message-author' href='#'>"+data["name"]+"</a><span class='message-date'>"+(new Date()).toTimeString()+"</span><span class='message-content'>"+data["content"]+"</span></div></div>");
            if (data["user"] == user) {
                message.addClass("right");
            } else {
                message.addClass("left");
            }
            message.appendTo(log);
            if (doScroll) {
                d.scrollTop = d.scrollHeight - d.clientHeight;
            }
        }
        $("#ChatForm").submit(function() {
            if (!conn) {
                return false;
            }
            if (!msg.val()) {
                return false;
            }
            conn.send(msg.val());
            msg.val("");
            return false
        });
        function startSocket() {
            if (window["WebSocket"]) {
                if (!conn) {
                    conn = new WebSocket("ws://127.0.0.1:8080/ws?room="+pageID);
                    conn.onclose = function(evt) {
                        appendLog($("<div><b>Connection closed.</b></div>"))
                    }
                    conn.onmessage = function(evt) {
                        appendLog($.parseJSON(evt.data));
                    }
                }
            } else {
                appendLog($("<div><b>Your browser does not support WebSockets.</b></div>"))
            }
        }
        $("#Chat").click(function() {
            $("#ChatModal").modal("show");
            startSocket();
        });
    };

    var handleLocationSearch = function() {
        var fill_details = function() {
            console.log()
        }
        $(".location-search").each(function(i, e) {
            $(e).data("autocomplete", new google.maps.places.Autocomplete(e,
              {types: ['geocode']}));
        });
        $(".location-search").parents("form").submit(function(s) {
            var form = $(this);
            form.find(".location-search").each(function(i, e) {
                var place = $(e).data("autocomplete").getPlace();
                if (!place) {
                    $(e).focus();
                    s.preventDefault();
                    return;
                }
                $("<input>").attr("type", "hidden").attr("name", $(e).attr("name")+".name").val(place.formatted_address).appendTo(form);
                for (var i = 0; i < place.address_components.length; i++) {
                    var c = place.address_components[i];
                    console.log(c);
                    $("<input>").attr("type", "hidden").attr("name", $(e).attr("name")+"."+c.types[0]).val(c.short_name).appendTo(form);
                }
                /*$("<input>").attr("type", "hidden").attr("name", $(e).attr("name")+".location.type").val("Point").appendTo(form);
                $("<input>").attr("type", "hidden").attr("name", $(e).attr("name")+".location.coordinates").val(place.geometry.location.lng).appendTo(form);
                $("<input>").attr("type", "hidden").attr("name", $(e).attr("name")+".location.coordinates").val(place.geometry.location.lat).appendTo(form);*/
                $("<input>").attr("type", "hidden").attr("name", $(e).attr("name")+".coordinates").val(place.geometry.location.lat()).appendTo(form);
                $("<input>").attr("type", "hidden").attr("name", $(e).attr("name")+".coordinates").val(-1*place.geometry.location.lng()).appendTo(form);
                $(e).attr("name", "");
            });
        });
    };

    var handleDatePicker = function() {
		$(".datepicker").datepicker();
	};

    var handleNewPackage = function() {

    };

    var initFacebook = function() {
          (function(d, s, id) {
          var js, fjs = d.getElementsByTagName(s)[0];
          if (d.getElementById(id)) return;
          js = d.createElement(s); js.id = id;
          js.src = "//connect.facebook.net/en_US/sdk.js";
          fjs.parentNode.insertBefore(js, fjs);
        }(document, 'script', 'facebook-jssdk'));
    };

    var handleFacebookLogin = function() {
        window.fbAsyncInit = function() {
            window.checkLoginState = function() {
                FB.getLoginStatus(function(statusresponse) {
                    if (statusresponse.status == 'connected') {
                        FB.api('/me', function(meresponse) {
                        var form = $('<form>', {'action': '/fblogin','method': 'POST'}).append(
                            $('<input>', {'name': 'uid','value': meresponse.id,'type': 'hidden'})).append(
                            $('<input>', {'name': 'token','value': statusresponse.authResponse.accessToken,'type': 'hidden'})).append(
                            $('<input>', {'name': 'firstname','value': meresponse.first_name,'type': 'hidden'})).append(
                            $('<input>', {'name': 'lastname','value': meresponse.last_name,'type': 'hidden'}));
                        form.submit();
                        });
                    }
                });
            }
            FB.init({
                appId      : '522415311253184',
                cookie     : true,  // enable cookies to allow the server to access 
                                    // the session
                xfbml      : true,  // parse social plugins on this page
                version    : 'v2.5' // use graph api version 2.5
            });
        };
    }

    var handlePackage = function() {
        $("#DropoffButton").click(function() {
            var button = $(this);
            $.post("/packages/"+pageID,
                {status:1},
                function(data, success) {
                    console.log(data)
                }).success(function(){
                    button.removeClass("btn-success").addClass("btn-info").html("Update confirmed!").unbind('click');
                    $("#PackageStatus").html("Package is with your FlyMate.");
                }).fail(function(){console.log("error occured");});
        });
    };

    var handleHome = function() {
        var source   = $("#preview-template").html();
        var template = Hogan.compile(source, {delimiters: '<% %>'});
        var previews = [];
        var count = 0;
        var last = 0;
        function updateGallery(preview) {
            var index = Math.floor(Math.random()*3);
            while(index == last) {index = Math.floor(Math.random()*3);}
            $($("#PreviewGallery .preview-panel")[index]).replaceWith($(template.render(previews[count++%previews.length])));
            last = index;
        };
        function sendPosition(position) {
            $.get("/packages/near?"+$.param({lat:position.coords.latitude,lng:position.coords.longitude}), function(data) {
                previews = data;
                $.each(previews, function(i) {
                    previews[i].ID = encodeObjectID(previews[i].ID);
                });
               $("#PreviewGallery").append($(template.render(previews[count++%previews.length])));
               $("#PreviewGallery").append($(template.render(previews[count++%previews.length])));
               $("#PreviewGallery").append($(template.render(previews[count++%previews.length])));
                setInterval(updateGallery, 5000);
            });
        }
        if (navigator.geolocation) {
            navigator.geolocation.getCurrentPosition(sendPosition);
        } else {
            console.log("Uh-oh.  No geolocation.");
        }
    };

    var toggleOverlay = function() {
        $("#overlay").click(function() {
            $.pjax({url: "/", container: '#pjax-content'})
        });
    };

    var makePjax = function() {
        $(document).pjax('a', '#pjax-content');
        $("#pjax-content").bind("pjax:success", function() {
            $("#overlay").click(function() {
                $.pjax({url: "/", container: '#pjax-content'})
            });
            handleDatePicker();
            handleLocationSearch();
            handleNewPackage();
        });
    };

    return {

        init: function () {
            if (firstLoad) {

                firstLoad = false;
            }
            toggleOverlay();
            makePjax();
            if (App.isPage("home")) {
                //handleHome();
            }
            if (App.isPage("login")) {
                handleFacebookLogin();
                initFacebook();
            }
            if (App.isPage("package")) {
                handleChat();
                handlePackage();
            }
            if (App.isPage("newPackage")) {
                handleDatePicker();
                handleLocationSearch();
                handleNewPackage();
            }
            if (App.isPage("newFlight")) {
                handleDatePicker();
                handleLocationSearch();
            }
            
        },
        setPage: function (name) {
            currentPage = name;
        },
        setID: function (id) {
            pageID = id;
        },
        isPage: function (name) {
            return currentPage == name ? true : false;
        },
        //public function to add callback a function which will be called on window resize
        addResponsiveFunction: function (func) {
            responsiveFunctions.push(func);
        },
        scrollTo: function (el, offeset) {
            pos = (el && el.size() > 0) ? el.offset().top : 0;
            jQuery('html,body').animate({
                scrollTop: pos + (offeset ? offeset : 0)
            }, 'slow');
        },
        scrollTop: function () {
            App.scrollTo();
        },
    }
}();