import { GetStatusCommand } from '@exodus/backend-contract'
import { createQueryKeys } from '@lukemorales/query-key-factory'

import { createGetQueryHook, errorHandler } from '../../tsq-helpers'

export const authQueryKeys = createQueryKeys('auth', {
    getAuthStatus: {
        queryKey: null
    }
})

export const useGetAuthStatus = createGetQueryHook({
    endpoint: GetStatusCommand.TSQ_url,
    responseSchema: GetStatusCommand.ResponseSchema,
    getQueryKey: () => authQueryKeys.getAuthStatus.queryKey,
    rQueryParams: {
        refetchOnMount: 'always',
        refetchOnWindowFocus: false,
        staleTime: 3000
    },
    errorHandler: (error: unknown) => errorHandler(error, 'Authentication Error')
})
