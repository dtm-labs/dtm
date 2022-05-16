const modules = import.meta.glob('../views/**/**.vue')
const components: IObject<() => Promise<typeof import('*.vue')>> = {
    LayoutHeader: (() => import('/@/layout/index.vue')) as unknown as () => Promise<typeof import('*.vue')>
}

Object.keys(modules).forEach(key => {
    const nameMatch = key.match(/^\.\.\/views\/(.+)\.vue/)
    if (!nameMatch) return
    if (nameMatch[1].includes('_Components')) return
    const indexMatch = nameMatch[1].match(/(.*)\/Index$/i)
    let name = indexMatch ? indexMatch[1] : nameMatch[1];
    [name] = name.split('/').splice(-1)
    components[name] = modules[key] as () => Promise<typeof import('*.vue')>
})

export {
    components
}
