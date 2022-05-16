import axios from 'axios'

const request = axios.create({
    baseURL: import.meta.env.VITE_APP_API_BASE_URL as string | undefined,
    timeout: 60000
})

export default request
