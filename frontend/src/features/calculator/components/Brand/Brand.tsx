import logoMark from '../../../../app/assets/logo-mark.png';
import { APP_NAME } from '../../../../shared/constants/app';
import styles from './Brand.module.css';

/**
 * The logo and name above the calculator; the name is the page's main heading.
 *
 * @returns The brand header.
 */
export function Brand() {
  return (
    <header className={styles.brand}>
      <img className={styles.logo} src={logoMark} alt="" />
      <h1 className={styles.name}>{APP_NAME}</h1>
    </header>
  );
}
