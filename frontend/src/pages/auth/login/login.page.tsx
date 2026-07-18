import { Badge, Box, Group, Image, Stack, Text, Title } from '@mantine/core'
import { GetStatusCommand } from '@exodus/backend-contract'
import { useEffect, useMemo } from 'react'

import { useGetAuthStatus } from '@shared/api/hooks'
import { LoginFormFeature } from '@features/auth/login-form'
import { parseColoredTextUtil } from '@shared/utils/misc'
import { Logo, Page } from '@shared/ui'

const BrandLogo = ({ logoUrl }: { logoUrl?: null | string }) => {
    if (!logoUrl) {
        return <Logo c="cyan" w="3rem" />
    }

    return (
        <Image
            alt="logo"
            fit="contain"
            src={logoUrl}
            style={{
                maxWidth: '40px',
                maxHeight: '40px',
                width: '40px',
                height: '40px'
            }}
        />
    )
}

const BrandTitle = ({ titleParts }: { titleParts: Array<{ color: string; text: string }> }) => {
    return (
        <Title ff="Unbounded" order={1} pos="relative">
            {titleParts.map((part, index) => (
                <Text
                    c={part.color || 'white'}
                    component="span"
                    fw="inherit"
                    fz="inherit"
                    inherit
                    key={index}
                    pos="relative"
                >
                    {part.text}
                </Text>
            ))}
        </Title>
    )
}

const stripBrandColorTags = (value: string) => value.replace(/\{[0-9a-fA-F]{3,8}\}/g, '').trim()

declare global {
    interface Window {
        __BRANDING__?: {
            title?: string
            description?: string
            logoUrl?: string | null
        }
    }
}

const applyBrandingMeta = (branding?: { title?: string; description?: string }) => {
    if (!branding) {
        return
    }

    const title = stripBrandColorTags(branding.title || '') || 'ProbeShield'
    document.title = title

    const description = branding.description || 'Authentication'
    const selector = 'meta[name="description"]'
    let meta = document.querySelector<HTMLMetaElement>(selector)
    if (!meta) {
        meta = document.createElement('meta')
        meta.name = 'description'
        document.head.appendChild(meta)
    }
    meta.content = description
}

export const LoginPage = () => {
    const { data: authStatus, isError } = useGetAuthStatus()

    useEffect(() => {
        const branding = authStatus?.branding ?? window.__BRANDING__
        applyBrandingMeta(branding)
    }, [authStatus?.branding])

    const titleParts = useMemo(() => {
        const title = authStatus?.branding?.title ?? window.__BRANDING__?.title
        if (title) {
            return parseColoredTextUtil(title)
        }

        return [
            { text: 'Probe', color: 'probeshield-logo-probe.6' },
            { text: 'Shield', color: 'probeshield-logo-shield.6' }
        ]
    }, [authStatus?.branding?.title])

    const logoUrl = authStatus?.branding?.logoUrl ?? window.__BRANDING__?.logoUrl
    const isPasswordEnabled = authStatus?.authentication?.password?.enabled ?? false

    return (
        <Page title="Authentication">
            <Stack align="center" gap="xs">
                <Group align="center" gap={4} justify="center">
                    <BrandLogo logoUrl={logoUrl} />
                    <BrandTitle titleParts={titleParts} />
                </Group>

                {isError && (
                    <Badge color="cyan" mt={10} size="lg" variant="filled">
                        Server is not responding. Check logs.
                    </Badge>
                )}

                {authStatus && isPasswordEnabled && (
                    <Box maw={800} p={30} w={{ base: 440, sm: 500, md: 500 }}>
                        <Stack gap="lg">
                            <LoginFormFeature />
                        </Stack>
                    </Box>
                )}
            </Stack>
        </Page>
    )
}

export default LoginPage
