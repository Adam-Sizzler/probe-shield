import { z } from 'zod'

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
      authentication: z.nullable(z.object({
        password: z.object({ enabled: z.boolean() })
      })),
      branding: z.object({
        title: z.string().nullable(),
        logoUrl: z.string().nullable()
      }),
      pageMeta: z.object({
        title: z.string(),
        description: z.string()
      })
    })
  })
}
