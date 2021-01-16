export class PlayerState {
    constructor(public pos: number, public lastPlay: Date) { }
}

class PlayerStateRepository {
    data: any;
    load() {
        const p = localStorage.getItem("player");
        if (p !== null) {
            this.data = JSON.parse(p);
        } else {
            this.data = {};
        }
    }
    store(stationId: string, start: string, pos: number) {
        this.data[`${stationId}-${start}`] = new PlayerState(pos, new Date());
        localStorage.setItem("player", JSON.stringify(this.data));
    }
    pos(stationId: string, start: string): number {
        const state = this.data[`${stationId}-${start}`];
        if (state !== undefined) {
            return state.pos;
        }
        return 0;
    }
}

const p = new PlayerStateRepository();
p.load();
export const playerStateRepository = p;