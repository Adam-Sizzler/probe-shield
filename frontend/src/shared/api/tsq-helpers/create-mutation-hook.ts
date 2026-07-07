import { useMutation, useQueryClient } from '@tanstack/react-query'
import { z } from 'zod'

import { createUrl, handleRequestError } from '../helpers'
import { instance } from '../axios'

type MutationParams<Response> = {
    onSuccess?: (data: Response, variables: any, context: unknown, queryClient: ReturnType<typeof useQueryClient>) => void
    onError?: (error: Error, variables: any, context: unknown, queryClient: ReturnType<typeof useQueryClient>) => void
    onSettled?: (data: Response | undefined, error: Error | null, variables: any, context: unknown, queryClient: ReturnType<typeof useQueryClient>) => void
}

export function createMutationHook<BodySchema extends z.ZodTypeAny, ResponseSchema extends z.ZodTypeAny>({
    endpoint,
    requestMethod,
    bodySchema,
    responseSchema,
    rMutationParams
}: {
    endpoint: string
    requestMethod: string
    bodySchema: BodySchema
    responseSchema: ResponseSchema
    rMutationParams?: MutationParams<z.infer<ResponseSchema>['response']>
}) {
    return (params?: { mutationFns?: MutationParams<z.infer<ResponseSchema>['response']> }) => {
        const queryClient = useQueryClient()

        const mutationFn = async ({ variables }: { variables?: z.infer<BodySchema> }) => {
            const url = createUrl(endpoint)

            return instance
                .request<z.infer<ResponseSchema>>({
                    method: requestMethod,
                    url,
                    data: bodySchema.parse(variables)
                })
                .then(async (response) => {
                    const result = await responseSchema.safeParseAsync(response.data)
                    if (!result.success) {
                        throw result.error
                    }
                    return result.data.response
                })
                .catch((error) => handleRequestError(error))
        }

        return useMutation<z.infer<ResponseSchema>['response'], Error, { variables?: z.infer<BodySchema> }>({
            mutationFn,
            onSuccess: (data, variables, context) => {
                rMutationParams?.onSuccess?.(data, variables, context, queryClient)
                params?.mutationFns?.onSuccess?.(data, variables, context, queryClient)
            },
            onError: (error, variables, context) => {
                rMutationParams?.onError?.(error, variables, context, queryClient)
                params?.mutationFns?.onError?.(error, variables, context, queryClient)
            },
            onSettled: (data, error, variables, context) => {
                rMutationParams?.onSettled?.(data, error, variables, context, queryClient)
                params?.mutationFns?.onSettled?.(data, error, variables, context, queryClient)
            }
        })
    }
}
