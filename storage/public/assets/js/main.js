const App = {
    data() {
        return {
            loading: true,
            index: {
                torrent: {
                    total: 0,
                    active: 0,
                    dead: 0,
                },
                peer: {
                    total: 0,
                    seeder: 0,
                    leecher: 0,
                },
                traffic: {
                    total: 0,
                    upload: 0,
                    download: 0,
                }
            },
            hot: []
        };
    },
    mounted: function () {
        fetch("/stats").then(res => res.json()).then((res) => {
            if (res.data.index) {
                const resIndexData = res.data.index;
                this.index.torrent.total = this.humanizeAmount(resIndexData.torrent.total);
                this.index.torrent.active = this.humanizeAmount(resIndexData.torrent.active);
                this.index.torrent.dead = this.humanizeAmount(resIndexData.torrent.dead);

                this.index.peer.total = this.humanizeAmount(resIndexData.peer.total);
                this.index.peer.seeder = this.humanizeAmount(resIndexData.peer.seeder);
                this.index.peer.leecher = this.humanizeAmount(resIndexData.peer.leecher);

                this.index.traffic.total = this.humanizeAmount(resIndexData.traffic.total, 'memory');
                this.index.traffic.upload = this.humanizeAmount(resIndexData.traffic.upload, 'memory');
                this.index.traffic.download = this.humanizeAmount(resIndexData.traffic.download, 'memory');
            }
            this.hot = res.data.hot || this.hot;
        }).catch((err) => {
            console.error(err);
        }).finally(() => {
            this.loading = false;
        })
    },
    methods: {
        handleDownloadBtnClick: function (index) {
            if (this.hot[index] && this.hot[index].info_hash) {
                const hashInfo = this.hot[index].info_hash;
                window.open('magnet:?xt=urn:btih:' + hashInfo + '&tr=' + window.location.href + 'announce')
            }
        },
        humanizeAmount: function (number, type = 'normal') {
            const typeMap = {
                normal: {
                    gap: 1000,
                    abbr: ['', 'K', 'M', 'B', 't', 'q', 'Q', 's', 'S']
                },
                memory: {
                    gap: 1024,
                    abbr: ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
                },
            };
            const gap = typeMap[type].gap;
            const abbrList = typeMap[type].abbr;

            let level = 1;
            while (level < abbrList.length) {
                if (number < Math.pow(gap, level)) {
                    break;
                }
                level = level + 1;
            }

            const fixedTo = level > 1 ? 2 : 0;
            return (number / Math.pow(gap, level - 1)).toFixed(fixedTo) + abbrList[level - 1];
        }
    }
};
const app = Vue.createApp(App);
app.use(ElementPlus);
app.mount("#app");