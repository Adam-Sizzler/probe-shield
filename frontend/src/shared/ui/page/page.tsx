import { forwardRef, ReactNode, useEffect } from 'react'
import { nprogress } from '@mantine/nprogress'
import { Box, BoxProps } from '@mantine/core'

interface PageProps extends BoxProps {
    children: ReactNode
    meta?: ReactNode
    title?: string
}

export const Page = forwardRef<HTMLDivElement, PageProps>(
    ({ children, meta, title: _title, ...other }, ref) => {
        useEffect(() => {
            nprogress.complete()
            return () => nprogress.start()
        }, [])

        return (
            <>
                {meta}
                <Box ref={ref} {...other}>
                    {children}
                </Box>
            </>
        )
    }
)
