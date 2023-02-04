import { ConfigEnv, UserConfigExport } from 'vite'
import path from 'path'
import vue from '@vitejs/plugin-vue'
import { createSvgIconsPlugin } from 'vite-plugin-svg-icons'
import Components from 'unplugin-vue-components/vite'
import { AntDesignVueResolver } from 'unplugin-vue-components/resolvers'
import { ViteEjsPlugin } from 'vite-plugin-ejs'
import dns from 'dns'

const setAlias = (alias: [string, string][]) =>
    alias.map((v) => {
        return { find: v[0], replacement: path.resolve(__dirname, v[1]) }
    })
// https://cn.vitejs.dev/config/server-options.html#server-host
dns.setDefaultResultOrder('verbatim')

export default ({ mode }: ConfigEnv): UserConfigExport => {
    return {
        resolve: {
            alias: setAlias([['/@', 'src']])
        },
        plugins: [
            vue(),
            createSvgIconsPlugin({
                iconDirs: [path.resolve(process.cwd(), 'src/icons')],
                symbolId: 'icon-[dir]-[name]'
            }),
            Components({
                dts: 'src/components.d.ts',
                resolvers: [AntDesignVueResolver()]
            }),
            ViteEjsPlugin({
                PUBLIC_PATH: mode !== 'development' ? 'PUBLIC-PATH-VARIABLE' : ''
            })
        ],
        experimental: {
            renderBuiltUrl(
                filename: string,
                {
                    hostType
                }: {
          hostId: string;
          hostType: 'js' | 'css' | 'html';
          type: 'asset' | 'public';
        }
            ) {
                if (hostType === 'js') {
                    return {
                        runtime: `window.__assetsPathBuilder(${JSON.stringify(filename)})`
                    }
                }

                return filename
            }
        },
        server: {
            host: 'localhost',
            port: 6789,
            base: 'admin',
            proxy: {
                '/api': {
                    changeOrigin: true,
                    target: 'http://localhost:36789'

                }
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
    }
}
