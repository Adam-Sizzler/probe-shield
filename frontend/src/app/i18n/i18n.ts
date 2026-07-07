import { initReactI18next } from 'react-i18next'
import i18n from 'i18next'

i18n.use(initReactI18next).init({
    lng: 'en',
    fallbackLng: 'en',
    supportedLngs: ['en'],
    defaultNS: 'exodus',
    ns: ['exodus'],
    resources: {
        en: {
            exodus: {
                'login-form': {
                    feature: {
                        username: 'Username',
                        password: 'Password',
                        'your-password': 'Your password',
                        'sign-in': 'Sign in'
                    }
                }
            }
        }
    },
    interpolation: {
        escapeValue: false
    },
    react: {
        useSuspense: false
    }
})

export default i18n
