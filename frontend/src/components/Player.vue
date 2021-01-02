<template>
  <div>
    <video ref="videoPlayer" class="video-js"></video>
  </div>
</template>

<script lang="ts">
import {
  defineComponent,
  reactive,
  onBeforeUnmount,
  ref,
  onMounted,
} from "vue";
import videojs from "video.js";

interface State {
  player: any;
}

export default defineComponent({
  name: "Player",
  props: {
    options: {
      type: Object,
    },
  },
  setup(props: any) {
    let player: any = null;
    const videoPlayer = ref<HTMLVideoElement>();
    onMounted(() => {
      console.log(videoPlayer);
      player = videojs(
        videoPlayer.value,
        props.options,
        function onPlayerReady() {
          console.log("onPlayerReady", player);
        }
      );
    });

    onBeforeUnmount(() => {
      player.dispose();
    });
    return {
      videoPlayer,
    };
  },
});
</script>
