import { useQuery } from '@tanstack/react-query'
import { z } from 'zod'

import { handleRequestError } from '../helpers'
import { instance } from '../axios'

type ResponsePayload<T extends z.ZodTypeAny> = z.infer<T> extends { response: infer R } ? R : never

export function createGetQueryHook<ResponseSchema extends z.ZodTypeAny>({
    endpoint,
    responseSchema,
    getQueryKey,
    rQueryParams
}: {
    endpoint: string
    responseSchema: ResponseSchema
    getQueryKey: () => readonly unknown[]
    rQueryParams?: Record<string, unknown>
    errorHandler?: (error: unknown) => void
}) {
    return () =>
        useQuery<ResponsePayload<ResponseSchema>>({
            queryKey: getQueryKey(),
            queryFn: async () =>
                instance
                    .get<z.infer<ResponseSchema>>(endpoint)
                    .then(async (response) => {
                        const result = await responseSchema.safeParseAsync(response.data)
                        if (!result.success) {
                            throw result.error
                        }
                        return result.data.response as ResponsePayload<ResponseSchema>
                    })
                    .catch((error) => handleRequestError(error)),
            ...(rQueryParams as any)
        })
}
