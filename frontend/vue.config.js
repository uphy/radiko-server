module.exports = {
    pages: {
        index: {
            entry: "src/main.ts",
            title: "Radiko Server",
        }
    },
    devServer: {
        port: 8081,
        host: '0.0.0.0',
        disableHostCheck: true,
        proxy: 'http://localhost:8080'
    },
    publicPath: "./"
};
