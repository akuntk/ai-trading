import { motion } from 'framer-motion'
import AnimatedSection from './AnimatedSection'
import { t, Language } from '../../i18n/translations'

function StepCard({ number, title, description, delay }: any) {
  return (
    <motion.div
      className="flex gap-6 items-start"
      initial={{ opacity: 0, x: -50 }}
      whileInView={{ opacity: 1, x: 0 }}
      viewport={{ once: true }}
      transition={{ delay }}
      whileHover={{ x: 10 }}
    >
      <motion.div
        className="flex-shrink-0 w-14 h-14 rounded-full flex items-center justify-center font-bold text-2xl"
        style={{
          background: 'var(--binance-yellow)',
          color: 'var(--brand-black)',
        }}
        whileHover={{ scale: 1.2, rotate: 360 }}
        transition={{ type: 'spring', stiffness: 260, damping: 20 }}
      >
        {number}
      </motion.div>
      <div>
        <h3
          className="text-2xl font-semibold mb-2"
          style={{ color: 'var(--brand-light-gray)' }}
        >
          {title}
        </h3>
        <p
          className="text-lg leading-relaxed"
          style={{ color: 'var(--text-secondary)' }}
        >
          {description}
        </p>
      </div>
    </motion.div>
  )
}

interface HowItWorksSectionProps {
  language: Language
}

export default function HowItWorksSection({
  language,
}: HowItWorksSectionProps) {
  return (
    <AnimatedSection id="how-it-works" backgroundColor="var(--brand-dark-gray)">
      <div className="max-w-7xl mx-auto">
        {/* 空内容，风险提示已移至页面底部 */}
      </div>
    </AnimatedSection>
  )
}
