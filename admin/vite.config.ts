import { ConfigEnv, UserConfigExport } from 'vite';
import path from 'path';
import vue from '@vitejs/plugin-vue';
import viteSvgIcons from 'vite-plugin-svg-icons';
import Components from 'unplugin-vue-components/vite';
import { AntDesignVueResolver } from 'unplugin-vue-components/resolvers'

const setAlias = (alias: [string, string][]) => alias.map((v) => {
    return { find: v[0], replacement: path.resolve(__dirname, v[1]) };
});

export default ({ }: ConfigEnv): UserConfigExport => {
    return {
        resolve: {
            alias: setAlias([['/@', 'src']]),
        },
        plugins: [
            vue(),
            viteSvgIcons({
                iconDirs: [path.resolve(process.cwd(), 'src/icons')],
                symbolId: 'icon-[dir]-[name]'
            }),
            Components({
                dts: 'src/components.d.ts',
                resolvers: [
                    AntDesignVueResolver()
                ]
            })
        ],
        server: {
            port: 6789,
            base: 'admin',
            proxy: {
                '/api': {
                    target: 'http://localhost:36789',
                },
            }
        },
        css: {
            postcss: {
                plugins: [
                    require('autoprefixer'),
                    require('tailwindcss'),
                    require('postcss-simple-vars'),
                    require('postcss-import')
                ]
            }
        }
    };
};
