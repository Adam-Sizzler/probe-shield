import { LoginCommand } from '@exodus/backend-contract'
import { notifications } from '@mantine/notifications'

import { setToken } from '@entities/auth/session-store'

import { createMutationHook } from '../../tsq-helpers'

export const AUTH_QUERY_KEY = 'auth'

export const useLogin = createMutationHook({
    endpoint: LoginCommand.TSQ_url,
    bodySchema: LoginCommand.RequestSchema,
    responseSchema: LoginCommand.ResponseSchema,
    requestMethod: LoginCommand.endpointDetails.REQUEST_METHOD,
    rMutationParams: {
        onSuccess: (data) => {
            setToken({ token: data.accessToken })
        },
        onError: (error) => {
            notifications.show({
                title: 'Login',
                message: error.message,
                color: 'red'
            })
        }
    }
})
