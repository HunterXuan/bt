const App = {
    data() {
        return {
            loading: true,
            trackerUrl: window.location.href + 'announce',
            index: {
                torrent: 0,
                peer: 0,
                traffic: 0
            },
            hot: []
        };
    },
    mounted: function () {
        fetch("/stats").then(res => res.json()).then((res) => {
            if (res.data.index) {
                const resIndexData = res.data.index;
                this.index.torrent = this.humanizeAmount(resIndexData.torrent);

                this.index.peer = this.humanizeAmount(resIndexData.peer);

                this.index.traffic = this.humanizeAmount(resIndexData.traffic, 'memory');
            }
            this.hot = res.data.hot || this.hot;
        }).catch((err) => {
            console.error(err);
        }).finally(() => {
            this.loading = false;
        })
    },
    methods: {
        goToGitHub: function () {
            window.open("https://github.com/HunterXuan/bt")
        },
        goToDockerHub: function () {
            window.open("https://hub.docker.com/r/hunterxuan/bt")
        },
        goToBlog: function () {
            window.open("https://hunterx.xyz")
        },
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