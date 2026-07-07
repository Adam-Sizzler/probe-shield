import axios from 'axios'

import { logoutEvents } from '@shared/emitters'

let authorizationToken = ''

export const instance = axios.create({
    baseURL: '',
    headers: {
        'Content-type': 'application/json',
        Accept: 'application/json'
    },
    withCredentials: true
})

instance.interceptors.request.use((config) => {
    if (authorizationToken) {
        config.headers.set('Authorization', `Bearer ${authorizationToken}`)
    }
    return config
})

export const setAuthorizationToken = (token: string) => {
    authorizationToken = token
}

instance.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response) {
            const responseStatus = error.response.status
            if (responseStatus === 401) {
                try {
                    logoutEvents.emit()
                } catch {
                    // no-op
                }
            }
        }
        return Promise.reject(error)
    }
)
