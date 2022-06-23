export interface IMenubar {
  menuList: Array<IMenubarList>
}

export interface ILayout {
  menubar: IMenubar
  status: IStatus
  dtmVersion: string
  globalError: string
}

export interface IStatus {
  isLoading: boolean
}

export interface IMenubarList {
  parentId?: number | string
  id?: number | string
  name: string
  path: string
  redirect?: string
  meta: {
    icon?: string
    title: string
    permission?: string[]
    activeMenu?: string
    hidden?: boolean
    alwaysShow?: boolean
  }
  component: (() => Promise<typeof import('*.vue')>) | string
  children?: Array<IMenubarList>
}
