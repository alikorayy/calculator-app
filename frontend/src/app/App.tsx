import { APP_NAME } from '../shared/constants/app';

/**
 * The application shell. It will render the calculator page; for now it only
 * proves that the toolchain builds, renders and tests a component.
 *
 * @returns The root element of the application.
 */
export function App() {
  return (
    <main>
      <h1>{APP_NAME}</h1>
    </main>
  );
}
