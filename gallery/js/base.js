$(document).ready(function() {

  var HOST = "http://localhost:3000"
  var currentImage
  var timerSwitch = false


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
  }, 1000)

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
          $("#imgCurrent").text("当前展示: " + imageName)
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
    $.get(HOST + "/api/images/contrast", function(data, status) {
      if (status == 'success') {
        var images = data.data
        var index = 0
        var ctimer = setInterval(function() {
          if (index >= 4) {
            clearInterval(ctimer)
          }
          $("#gallery").attr("src", images[index])
          index ++
        }, 3000)

      }
    })
  }

  $("#play").click(function() {
    startAction()
  });

  $("#contrast").click(function() {
    showContrast()
  });

})
