import { z } from 'zod'

export declare const OAUTH2_PROVIDERS: {
  readonly GITHUB: 'github'
  readonly GOOGLE: 'google'
  readonly YANDEX: 'yandex'
}

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
      isRegisterAllowed: boolean
      authentication: null | {
        passkey: { enabled: boolean }
        tgAuth: { enabled: boolean; botId: number | null }
        oauth2: { providers: Record<string, boolean> }
        password: { enabled: boolean }
      }
      branding: { title: string | null; logoUrl: string | null }
    }
  }>
}
export declare namespace GetStatusCommand {
  type Response = z.infer<typeof GetStatusCommand.ResponseSchema>
}
