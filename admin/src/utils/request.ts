import axios from 'axios'

const request = axios.create({
    timeout: 60000
})

export default request
