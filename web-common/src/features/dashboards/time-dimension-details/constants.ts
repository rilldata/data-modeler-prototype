export const ROW_HEIGHT = 24;
export const FILTER_OVERFLOW_WIDTH = 24;
export const HEADER_HEIGHT = 36;

export const PieChart = `<svg height="1em" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"> <path  d="M14.7497 1.01468C14.4741 0.998446 14.25 1.22384 14.25 1.49998V9.49998L22.25 9.49998C22.5261 9.49998 22.7515 9.27593 22.7353 9.00027C22.6796 8.0549 22.4663 7.12426 22.103 6.24717C21.6758 5.21591 21.0497 4.27887 20.2604 3.48957C19.4711 2.70027 18.5341 2.07417 17.5028 1.647C16.6257 1.2837 15.6951 1.07035 14.7497 1.01468Z"  fill="currentColor" /> <path  fill-rule="evenodd"  clip-rule="evenodd"  d="M12.5 4.49998C12.5 3.82296 11.9419 3.22085 11.2035 3.26706C9.66504 3.36333 8.17579 3.86466 6.88876 4.72462C5.44983 5.68608 4.32832 7.05264 3.66606 8.6515C3.00379 10.2504 2.83051 12.0097 3.16813 13.707C3.50575 15.4044 4.33911 16.9635 5.56282 18.1872C6.78653 19.4109 8.34563 20.2442 10.043 20.5819C11.7403 20.9195 13.4996 20.7462 15.0985 20.0839C16.6973 19.4217 18.0639 18.3001 19.0254 16.8612C19.8853 15.5742 20.3867 14.0849 20.4829 12.5465C20.5291 11.8081 19.927 11.25 19.25 11.25H12.5V4.49998ZM7.72212 5.97182C8.64349 5.35618 9.68996 4.96237 10.7797 4.8152C10.8973 4.79932 11 4.89211 11 5.01075V12.75H18.7392C18.8579 12.75 18.9507 12.8527 18.9348 12.9703C18.7876 14.06 18.3938 15.1065 17.7782 16.0279C16.9815 17.2201 15.8492 18.1494 14.5245 18.6981C13.1997 19.2468 11.742 19.3904 10.3356 19.1107C8.92924 18.8309 7.63741 18.1404 6.62348 17.1265C5.60955 16.1126 4.91905 14.8207 4.63931 13.4144C4.35957 12.008 4.50314 10.5503 5.05188 9.22552C5.60061 7.90076 6.52986 6.76846 7.72212 5.97182Z"  fill="currentColor" /></svg>`;

export const MeasureArrow = (isAscending) => `<svg
width="1em"
height="1em"
viewBox="0 0 24 24"
fill="currentColor"
transform="${isAscending ? "scale(1 -1)" : "scale(1 1)"}"
xmlns="http://www.w3.org/2000/svg"
>
<path
  d="M18.3672 13.7775C18.8016 13.7775 19.0294 14.2935 18.7366 14.6145L12.3694 21.595C12.1711 21.8124 11.8289 21.8124 11.6306 21.595L5.26341 14.6145C4.97063 14.2935 5.19836 13.7775 5.63282 13.7775H10.4023L10.4023 2.5C10.4023 2.22386 10.6262 2 10.9023 2L13.4379 2C13.714 2 13.9379 2.22386 13.9379 2.5L13.9379 13.7775L18.3672 13.7775Z"
  fill="currentColor"
/>
</svg>`;

export const CheckCirlce = (className) => `<svg
height="16px"
width="16px"
viewBox="0 0 24 24"
fill="none"
xmlns="http://www.w3.org/2000/svg"
>
<path
  fill-rule="evenodd"
  clip-rule="evenodd"
  d="M12 22C17.5228 22 22 17.5228 22 12C22 6.47715 17.5228 2 12 2C6.47715 2 2 6.47715 2 12C2 17.5228 6.47715 22 12 22ZM10.4977 18.1532L18.5873 8.89285L16.8928 7.4126L10.3986 14.8467L7.01172 11.409L5.40894 12.9881L10.4977 18.1532Z"
  class=${className}
  fill="currentColor"
/>
</svg>`;

export const ExcludeIcon = `<svg
height="18px"
viewBox="0 0 24 24"
fill="none"
xmlns="http://www.w3.org/2000/svg"
>
<path
  fill="currentColor"
  fill-rule="evenodd"
  clip-rule="evenodd"
  d="M7.61339 5.90083C7.7354 5.77882 7.93322 5.77882 8.05523 5.90083L11.7791 9.62468C11.9011 9.74669 12.0989 9.74669 12.2209 9.62468L15.9448 5.90083C16.0668 5.77882 16.2646 5.77882 16.3866 5.90083L17.9085 7.42272C18.0305 7.54473 18.0305 7.74255 17.9085 7.86456L14.1846 11.5884C14.0626 11.7104 14.0626 11.9082 14.1846 12.0302L17.9085 15.7541C18.0305 15.8761 18.0305 16.0739 17.9085 16.1959L16.3866 17.7178C16.2646 17.8398 16.0668 17.8398 15.9448 17.7178L12.2209 13.994C12.0989 13.872 11.9011 13.872 11.7791 13.994L8.05523 17.7178C7.93322 17.8398 7.7354 17.8398 7.61339 17.7178L6.09151 16.1959C5.9695 16.0739 5.9695 15.8761 6.09151 15.7541L9.81536 12.0302C9.93737 11.9082 9.93737 11.7104 9.81536 11.5884L6.09151 7.86456C5.9695 7.74255 5.9695 7.54473 6.09151 7.42272L7.61339 5.90083Z"
/>
</svg>`;
