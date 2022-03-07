<template>
<div>
  <el-row>
    <el-col :span="6">
      <div v-if="started">
        <el-button
          @click="stopAction()" type="warning">停止播放
        </el-button>

        <div>
          <el-link type="primary">当前播放: {{ imageName }}</el-link>
        </div>
      </div>
      <div v-else>
        <el-button
          @click="startAction()" type="primary">开始播放
        </el-button>
      </div>
    </el-col>

    <el-col :span="12">

      <div :visible="started">
        <img :src="url" width="730" height="730">
      </div>


    <el-card class="box-card">
      <div slot="header" class="clearfix">
        <span>{{ imageCount }}张图片准备就绪</span>
      </div>
      <div v-for="name in imagesContraction" :key="name" class="text item">
        {{ name }}
      </div>
      <div>...</div>
    </el-card>

<!--     <div class="demo-image__preview">
      <el-image-viewer
      v-if="showImageView"
      :on-close="closeView"
      :url-list="[url]">
      </el-image-viewer>
    </div> -->

    </el-col>
  </el-row>

</div>
</template>

<script>
import { getImages, getImage, imageReady } from '@/js/requests'
// import ElImageViewer from 'element-ui/packages/image/src/image-viewer'

export default {
  // components: { ElImageViewer },
  data() {
    return {
      images: [],
      imagesContraction: [],
      imageCount: 0,
      index: 0,
      url: "",
      showImageView: false,
      timer: "",
      timerSwitch: false,
      started: false,
      imageName: ""
    }
  },
  created () {
    this.refresh()
  },
  methods: {
    refresh() {
      getImages().then(res => {
        if (res.data.code == 0) {
          this.$message({ message: '获取图片列表成功', type: 'success' });
          this.images = res.data.data
          this.imagesContraction = res.data.data.slice(0,5)
          this.imageCount = res.data.data.length

        } else {
          this.$message({ message: '获取图片列表失败', type: 'warning' });

        }
      })
    },
    getImageURL() {
      getImage().then(res => {
        if (res.data.code == 0) {
          if (res.data.data == "") {
            this.showImageView = false
            this.url = ""
            this.timerSwitch = false
            alert("图片展示完毕")
          }

          var update
          if (this.url == "") {
            update = true
          } else if (this.url != res.data.data) {
            update = true
          } else {
            update = false
          }
          if (update) {
            this.url = res.data.data
            this.showImageView = true
            imageReady()

            // http://localhost:3000/images/p_1.jpg
            var parts = this.url.split('/')
            this.imageName = parts[parts.length - 1]
          }

        } else {
          this.$message({ message: '获取图片地址失败', type: 'warning' });
          this.timerSwitch = true
        }
      })
    },
    startAction() {
      this.started = true
      this.timerSwitch = true
      this.getImageURL()
    },
    stopAction() {
      this.started = false
      this.showImageView = false
      this.url = ""
      this.timerSwitch = false
    },
    timerAction() {
      if (this.timerSwitch) {
        this.getImageURL()
      }
    }
  },
  mounted() {
    this.timer = setInterval(this.timerAction, 500);
  },
  beforeDestroy() {
    clearInterval(this.timer);
  }
}
</script>

<style>
  .el-row {
    margin-top: 60px;
  }
</style>