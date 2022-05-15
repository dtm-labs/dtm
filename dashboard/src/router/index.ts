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
    redirect: '/dashboard/global-transactions/all',
    component: Components['LayoutHeader'],
    meta: { title: 'Dashboard', activeMenu: '/dashboard' },
    children: [
      {
        name: 'Nodes',
        path: '/dashboard/nodes',
        component: Components['LayoutMain'],
        meta: { title: 'Nodes' },
        children: [
          {
            name: 'LivingNodes',
            path: '/dashboard/nodes/living',
            component: Components['LivingNodes'],
            meta: { title: 'Living Nodes' },
          }
        ]
      }, {
        name: 'GlobalTransactions',
        path: '/dashboard/global-transactions',
        component: Components['LayoutMain'],
        meta: { title: 'Global Transactions' },
        children: [
          {
            name: 'AllTransactions',
            path: '/dashboard/global-transactions/all',
            component: Components['AllTransactions'],
            meta: { title: 'All Transactions' },
          }, {
            name: 'UnfinishedTransactions',
            path: '/dashboard/global-transactions/unfinished',
            component: Components['UnfinishedTransactions'],
            meta: { title: 'Unfinished Transactions' },
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
