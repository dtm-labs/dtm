import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { IMenubarList } from '../type/store/layout';
import { components } from './asyncRouter';

const Components: IObject<() => Promise<typeof import('*.vue')>> = Object.assign({}, components, {
  LayoutHeader: (() => import('/@/layout/index.vue')) as unknown as () => Promise<typeof import('*.vue')>,
  LayoutMain: (() => import('/@/layout/aside.vue')) as unknown as () => Promise<typeof import('*.vue')>
})

export const allowRouter: Array<IMenubarList> = [
    {
    name: 'Dashboard',
    path: '/',
    redirect: '/dashboard/sub1/1',
    component: Components['LayoutHeader'],
    meta: { title: '首页', activeMenu: '/dashboard' },
    children: [
      {
        name: 'SubNav1',
        path: '/dashboard/sub1',
        component: Components['LayoutMain'],
        meta: { title: '子导航1' },
        children: [
            {
                name: 'Page1',
                path: '/dashboard/sub1/1',
                component: Components['DashboardPage1'],
                meta: { title: '子页面1' },
            },  {
                name: 'Page2',
                path: '/dashboard/sub1/2',
                component: Components['DashboardPage2'],
                meta: { title: '子页面2' },
            }
        ]
      }, {
        name: 'SubNav2',
        path: '/dashboard/sub2',
        component: Components['LayoutMain'],
        meta: { title: '子导航2' },
        children: [
            {
                name: 'Page21',
                path: '/dashboard/sub2/1',
                component: Components['DashboardPage1'],
                meta: { title: '子页面21' },
            },  {
                name: 'Page22',
                path: '/dashboard/sub2/2',
                component: Components['DashboardPage2'],
                meta: { title: '子页面22' },
            }
        ]
      }
    ]
  },   {
    name: 'Nav2',
    path: '/nav2',
    redirect: '/nav2/sub2/1',
    component: Components['LayoutHeader'],
    meta: { title: '导航2', activeMenu: '/nav2' },
    children: [
      {
        name: 'Nav2Sub1',
        path: '/nav2/sub1',
        component: Components['LayoutMain'],
        meta: { title: '子导航1' },
        children: [
            {
                name: 'Nav2Page1',
                path: '/nav2/sub1/1',
                component: Components['DashboardPage1'],
                meta: { title: '子页面1' },
            },  {
                name: 'Nav2Page2',
                path: '/nav2/sub1/2',
                component: Components['DashboardPage2'],
                meta: { title: '子页面2' },
            }
        ]
      }, {
        name: 'Nav2Sub2',
        path: '/nav2/sub2',
        component: Components['LayoutMain'],
        meta: { title: '子导航2' },
        children: [
            {
                name: 'Nav2Page21',
                path: '/nav2/sub2/1',
                component: Components['DashboardPage1'],
                meta: { title: '子页面21' },
            },  {
                name: 'Nav2Page22',
                path: '/nav2/sub2/2',
                component: Components['DashboardPage2'],
                meta: { title: '子页面22' },
            }
        ]
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes: allowRouter as RouteRecordRaw[]
})

export default router
