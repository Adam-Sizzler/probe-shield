import { isAxiosError } from 'axios'
import { ZodError } from 'zod'

export function handleRequestError(error: unknown) {
    if (isAxiosError(error)) {
        const errorData = error.response?.data
        const message =
            typeof errorData === 'string'
                ? errorData
                : typeof errorData?.message === 'string'
                  ? errorData.message
                  : typeof errorData?.error === 'string'
                    ? errorData.error
                    : typeof errorData?.error?.message === 'string'
                      ? errorData.error.message
                      : 'Request failed'
        const enhancedError = new Error(message)
        enhancedError.cause = errorData
        throw enhancedError
    }

    if (error instanceof ZodError) {
        throw new Error('Invalid server response')
    }

    throw error
}
