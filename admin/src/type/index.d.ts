export { }
declare global {
  interface IObject<T> {
    [index: string]: T
  }
  interface ImportMetaEnv {
    VITE_APP_TITLE: string
    VITE_PORT: number
    VITE_PROXY: string
    VITE_ADMIN_VERSION: string
  }
  interface ITable<T = any> {
    data: Array<T>
    next_position: number,
    size: number
  }
  interface Window {
    basePath: string;
  }
}
