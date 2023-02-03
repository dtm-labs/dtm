import axios from 'axios'

const request = axios.create({
    baseURL: window.basePath || '',
    timeout: 60000
})

export default request
