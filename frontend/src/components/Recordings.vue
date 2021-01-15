<template>
  <div>
    <div
      class="recordings"
      v-for="(recordings, i) in recordingsGroupByDate"
      :key="i"
    >
      <div class="recordings-date">{{ recordings.date.toString() }}</div>
      <div
        class="recording"
        v-for="(recording, i) in recordings.recordings"
        :key="i"
      >
        <div class="recording-title">
          <router-link :to="routerLink(recording)">{{
            recording.title
          }}</router-link>
        </div>
        <div class="recording-date">
          {{ recording.start.toLocaleString() }} -
          {{ recording.end.toLocaleString() }}
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent } from "vue";
import { Recording } from "../api";

type Props = {
  recordings: Recording[];
};

class RecordingsByDate {
  constructor(public date: MonthDate, public recordings: Recording[]) {}
}

class MonthDate {
  constructor(public year: number, public month: number, public date: number) {}
  equals(other: MonthDate): boolean {
    return (
      this.year === other.year &&
      this.month === other.month &&
      this.date === other.date
    );
  }
  toString(): string {
    return `${this.year}/${this.month}/${this.date}`;
  }
}

export default defineComponent({
  name: "Recordings",
  props: {
    recordings: {
      type: Array,
      default: () => [],
    },
  },
  setup<Props>(props: any) {
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
      recordingsGroupByDate: computed(() => {
        let result: RecordingsByDate[] = [];
        let current: RecordingsByDate | null = null;
        const sorted: Recording[] = [...props.recordings];
        sorted.sort((a: any, b: any) => {
          return b.start - a.start;
        });

        sorted.forEach((recording: any) => {
          const date = new MonthDate(
            recording.start.getFullYear(),
            recording.start.getMonth() + 1,
            recording.start.getDate()
          );
          if (current === null || !current.date.equals(date)) {
            current = new RecordingsByDate(date, []);
            result.push(current);
          }
          current.recordings.push(recording);
        });
        return result;
      }),
    };
  },
});
</script>

<style scoped>
.recordings {
  background-color: #e6c3f8;
  margin-bottom: 0.5rem;
}
.recordings-date {
  font-size: 1.5rem;
  display: block;
  padding: 0 0.5rem;
  margin: 0;
  background-color: #8f00d8;
  color: #dbdbdb;
}
.recording {
  padding: 0.5rem;
}
.recording-title {
  font-size: 1rem;
}
.recording-title > a {
  color: #333333;
}
.recording-date {
  font-size: 0.5rem;
}
</style>
