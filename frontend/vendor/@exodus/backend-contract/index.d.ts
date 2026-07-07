import { z } from 'zod'

export declare const LoginCommand: {
  url: string
  TSQ_url: string
  endpointDetails: { REQUEST_METHOD: 'post' }
  RequestSchema: z.ZodObject<{ username: z.ZodString; password: z.ZodString }>
  ResponseSchema: z.ZodObject<{ response: z.ZodObject<{ accessToken: z.ZodString }> }>
}
export declare namespace LoginCommand {
  type Response = z.infer<typeof LoginCommand.ResponseSchema>
}

export declare const GetStatusCommand: {
  url: string
  TSQ_url: string
  endpointDetails: { REQUEST_METHOD: 'get' }
  ResponseSchema: z.ZodType<{
    response: {
      isLoginAllowed: boolean
      authentication: null | {
        password: { enabled: boolean }
      }
      branding: { title: string | null; logoUrl: string | null }
      pageMeta: { title: string; description: string }
    }
  }>
}
export declare namespace GetStatusCommand {
  type Response = z.infer<typeof GetStatusCommand.ResponseSchema>
}
