import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import useBaseUrl from '@docusaurus/useBaseUrl';
import Head from '@docusaurus/Head';
import siteConfig from '@generated/docusaurus.config';

import styles from './index.module.css';

const version = siteConfig.customFields?.version as string || 'dev';

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  const logoUrl = useBaseUrl('/img/logo.svg');
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className={styles.alphaBanner}>
        Early Alpha · {version}
      </div>
      <div className="container">
        <img
          src={logoUrl}
          alt="Drift"
          className={styles.heroLogo}
        />
        <p className="hero__subtitle">{siteConfig.tagline}</p>
        <p className={styles.heroFeature}>Native GPU rendering · No VM overhead</p>
        <div className={styles.buttons}>
          <Link
            className="button button--secondary button--lg"
            to="/docs/intro">
            Documentation
          </Link>
          <Link
            className="button button--secondary button--lg"
            to="/docs/category/api-reference">
            API Reference
          </Link>
          <Link
            className="button button--secondary button--lg"
            href="https://github.com/go-drift/drift">
            GitHub
          </Link>
        </div>
      </div>
    </header>
  );
}

export default function Home(): JSX.Element {
  const {siteConfig} = useDocusaurusContext();

  return (
    <>
      <Head>
        <title>{`${siteConfig.title} - ${siteConfig.tagline}`}</title>
        <meta name="description" content="Drift is a cross-platform mobile UI framework in Go." />
      </Head>
      <HomepageHeader />
    </>
  );
}
