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
            this.index = res.data.index || this.index;
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
        }
    }
};
const app = Vue.createApp(App);
app.use(ElementPlus);
app.mount("#app");