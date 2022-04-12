import { ConfigEnv, UserConfigExport } from 'vite';
import path from 'path';
import vue from '@vitejs/plugin-vue';
import viteSvgIcons from 'vite-plugin-svg-icons';

const setAlias = (alias: [string, string][]) => alias.map((v) => {
    return { find: v[0], replacement: path.resolve(__dirname, v[1]) };
});

export default ({}: ConfigEnv): UserConfigExport => {
    return {
        resolve: {
            alias: setAlias([['/@', 'src']]),
        },
        plugins: [
            vue(),
            viteSvgIcons({
                iconDirs: [path.resolve(process.cwd(), 'src/icons')],
                symbolId: 'icon-[dir]-[name]'
            })
        ],
        css: {
            postcss: {
                plugins: [
                    require('autoprefixer'),
                    require('postcss-simple-vars'),
                    require('postcss-import')
                ]
            }
        }
    };
};
