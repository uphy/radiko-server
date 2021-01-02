<template>
  <div>
    <table>
      <thead>
        <tr>
          <th>Title</th>
          <th>Start</th>
          <th>End</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(recording, i) in recordings" :key="i">
          <td>{{ recording.title }}</td>
          <td>{{ recording.start }}</td>
          <td>{{ recording.end }}</td>
          <td><router-link :to="routerLink(recording)">Play</router-link></td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

export default defineComponent({
  name: "Recordings",
  props: {
    recordings: {
      type: Array,
      default: () => [],
    },
  },
  setup(props: any) {
    const p = (s: string, len: number): string => {
      for (let i = s.length; i < len; i++) {
        s = "0" + s;
      }
      return s;
    };
    return {
      routerLink(recording: any): string {
        const start: Date = recording.start;
        return `/recordings/${recording.stationId}/${start.getFullYear()}${p(
          (start.getMonth() + 1).toString(),
          2
        )}${p(start.getDate().toString(), 2)}${p(
          start.getHours().toString(),
          2
        )}${p(start.getMinutes().toString(), 2)}${p(
          start.getSeconds().toString(),
          2
        )}`;
      },
    };
  },
});
</script>

<style scoped></style>
