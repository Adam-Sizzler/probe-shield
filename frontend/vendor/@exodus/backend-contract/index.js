import { z } from 'zod'

export const OAUTH2_PROVIDERS = {
  GITHUB: 'github',
  GOOGLE: 'google',
  YANDEX: 'yandex'
}

export const LoginCommand = {
  url: '/api/auth/login',
  TSQ_url: '/api/auth/login',
  endpointDetails: { REQUEST_METHOD: 'post' },
  RequestSchema: z.object({
    username: z.string(),
    password: z.string()
  }),
  ResponseSchema: z.object({
    response: z.object({
      accessToken: z.string()
    })
  })
}

export const GetStatusCommand = {
  url: '/api/auth/status',
  TSQ_url: '/api/auth/status',
  endpointDetails: { REQUEST_METHOD: 'get' },
  ResponseSchema: z.object({
    response: z.object({
      isLoginAllowed: z.boolean(),
      isRegisterAllowed: z.boolean(),
      authentication: z.nullable(z.object({
        passkey: z.object({ enabled: z.boolean() }),
        tgAuth: z.object({ enabled: z.boolean(), botId: z.number().nullable() }),
        oauth2: z.object({ providers: z.record(z.string(), z.boolean()) }),
        password: z.object({ enabled: z.boolean() })
      })),
      branding: z.object({
        title: z.string().nullable(),
        logoUrl: z.string().nullable()
      })
    })
  })
}
