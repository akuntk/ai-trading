import { useState } from 'react'
import { motion } from 'framer-motion'
import { ArrowRight } from 'lucide-react'
import HeaderBar from '../components/HeaderBar'
import HeroSection from '../components/landing/HeroSection'
import AboutSection from '../components/landing/AboutSection'
import FeaturesSection from '../components/landing/FeaturesSection'
import HowItWorksSection from '../components/landing/HowItWorksSection'
import CommunitySection from '../components/landing/CommunitySection'
import AnimatedSection from '../components/landing/AnimatedSection'
import LoginModal from '../components/landing/LoginModal'
import FooterSection from '../components/landing/FooterSection'
import { useAuth } from '../contexts/AuthContext'
import { useLanguage } from '../contexts/LanguageContext'
import { t } from '../i18n/translations'

export function LandingPage() {
  const [showLoginModal, setShowLoginModal] = useState(false)
  const { user, logout } = useAuth()
  const { language, setLanguage } = useLanguage()
  const isLoggedIn = !!user

  console.log('LandingPage - user:', user, 'isLoggedIn:', isLoggedIn)
  return (
    <>
      <HeaderBar
        onLoginClick={() => setShowLoginModal(true)}
        isLoggedIn={isLoggedIn}
        isHomePage={true}
        language={language}
        onLanguageChange={setLanguage}
        user={user}
        onLogout={logout}
        onPageChange={(page) => {
          console.log('LandingPage onPageChange called with:', page)
          if (page === 'competition') {
            window.location.href = '/competition'
          } else if (page === 'traders') {
            window.location.href = '/traders'
          } else if (page === 'trader') {
            window.location.href = '/dashboard'
          }
        }}
      />
      <div
        className="min-h-screen px-4 sm:px-6 lg:px-8"
        style={{
          background: 'var(--brand-black)',
          color: 'var(--brand-light-gray)',
        }}
      >
        <HeroSection language={language} />
        {/* <AboutSection language={language} /> */}
        <FeaturesSection language={language} />
        <HowItWorksSection language={language} />
        {/* <CommunitySection /> */}

        {/* CTA */}
        <AnimatedSection backgroundColor="var(--panel-bg)">
          <div className="max-w-4xl mx-auto text-center">
            <motion.h2
              className="text-5xl font-bold mb-6"
              style={{ color: 'var(--brand-light-gray)' }}
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
            >
              {t('readyToDefine', language)}
            </motion.h2>
            <motion.p
              className="text-xl mb-12"
              style={{ color: 'var(--text-secondary)' }}
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: 0.1 }}
            >
              {t('startWithCrypto', language)}
            </motion.p>
            <div className="flex flex-wrap justify-center gap-4">
              <motion.button
                onClick={() => setShowLoginModal(true)}
                className="flex items-center gap-2 px-10 py-4 rounded-lg font-semibold text-lg"
                style={{
                  background: 'var(--brand-yellow)',
                  color: 'var(--brand-black)',
                }}
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
              >
                {t('getStartedNow', language)}
                <motion.div
                  animate={{ x: [0, 5, 0] }}
                  transition={{ duration: 1.5, repeat: Infinity }}
                >
                  <ArrowRight className="w-5 h-5" />
                </motion.div>
              </motion.button>
              {/* <motion.a
                href="https://github.com/akuntk/ai-trading"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-2 px-10 py-4 rounded-lg font-semibold text-lg"
                style={{
                  background: 'transparent',
                  color: 'var(--brand-light-gray)',
                  border: '2px solid var(--brand-yellow)',
                }}
                whileHover={{
                  scale: 1.05,
                  backgroundColor: 'rgba(240, 185, 11, 0.1)',
                }}
                whileTap={{ scale: 0.95 }}
              >
                {t('viewSourceCode', language)}
              </motion.a> */}
            </div>
          </div>
        </AnimatedSection>

        {showLoginModal && (
          <LoginModal
            onClose={() => setShowLoginModal(false)}
            language={language}
          />
        )}

        {/* 风险提示 - 移至页面底部 */}
        <div className="max-w-4xl mx-auto px-6 py-8">
          <motion.div
            className="p-6 rounded-xl flex items-start gap-4"
            style={{
              background: 'rgba(246, 70, 93, 0.1)',
              border: '1px solid rgba(246, 70, 93, 0.3)',
            }}
            initial={{ opacity: 0, scale: 0.9 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            whileHover={{ scale: 1.02 }}
          >
            <div
              className="w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0"
              style={{ background: 'rgba(246, 70, 93, 0.2)', color: '#F6465D' }}
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="w-6 h-6"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0Z" />
                <line x1="12" x2="12" y1="9" y2="13" />
                <line x1="12" x2="12.01" y1="17" y2="17" />
              </svg>
            </div>
            <div>
              <div className="font-semibold mb-2" style={{ color: '#F6465D' }}>
                {t('importantRiskWarning', language)}
              </div>
              <p className="text-sm" style={{ color: 'var(--text-secondary)' }}>
                {t('riskWarningText', language)}
              </p>
            </div>
          </motion.div>
        </div>
        {/* <FooterSection language={language} /> */}
      </div>
    </>
  )
}
