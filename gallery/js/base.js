$(document).ready(function() {

  var HOST = "http://localhost:3000"
  var currentImage
  var timerSwitch = false
  var contrastSwitch = false
  var delaySwitch = false


  $.get(HOST + "/api/images", function(data,status){
    if (status == 'success') {
      $("#imgCount").text(data.data.length + "张图片准备就绪")
    } else {
      alert("图片列表请求失败")
    }
  });

  var timer = setInterval(function() {
    if (timerSwitch == true) {
      getImage()
    }
  }, 500)

  function imageReady() {
    $.get(HOST + "/api/action?type=2", function(data, status) {
      if (status != 'success') {
        alert("图片就绪请求失败")
      }
    })
  };

  function getImage() {
    $.get(HOST + "/api/image", function(data, status) {
      if (status == 'success') {
        if (data.data == "") {
          timerSwitch = false

        } else if (currentImage != data.data) {
          $("#gallery").attr("src", data.data)
          currentImage = data.data
          var parts = currentImage.split('/')
          imageName = parts[parts.length - 1]
          $("#warning").text("当前展示: " + imageName)
          imageReady()
          $("#gallery").attr("src", data.data)
        }
      } else {
        alert("图片地址请求失败")
      }
    })
  }

  function startAction() {
    timerSwitch = true
    getImage()
  }

  function showContrast() {
    if (contrastSwitch == true) {
      return
    }
    contrastSwitch = true

    $.get(HOST + "/api/images/contrast", function(data, status) {
      if (status == 'success') {
        var images = data.data
        var index = 0
        var ctimer = setInterval(function() {
          if (index >= 3) {
            clearInterval(ctimer)
            contrastSwitch = false
            $("#warning").text("黑白图展示完毕，请刷新界面")
          }
          $("#gallery").attr("src", images[index])
          index ++
        }, 3000)
      }
    })
  }

  function delayAction() {
    if (delaySwitch == true) {
      return
    }
    delaySwitch = true

    var delayTime = Number($("#delayTime").val())
    var dtimer = setInterval(function() {
      if (delayTime <= 0) {
        clearInterval(dtimer)
        delaySwitch = false
        startAction()
      }
      $("#warning").text("散斑投影将在: " + delayTime + " 秒后开始")
      delayTime = delayTime - 1
    }, 1000)
  }

  $("#play").click(function() {
    startAction()
  });

  $("#contrast").click(function() {
    showContrast()
  });

  $("#delay").click(function() {
    delayAction()
  });

})
