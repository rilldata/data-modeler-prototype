interface Link {
  text: string;
  href: string;
}

export interface NotificationMessage {
  type?: "default" | "success" | "error";
  message: string;
  detail?: string;
  link?: Link;
  options?: NotificationOptions;
}

interface NotificationOptions {
  persisted?: boolean;
}

export interface BannerMessage {
  type: "default" | "success" | "info" | "warning" | "error";
  message: string;
  includesHtml?: boolean;

  iconType: "none" | "alert" | "check" | "sleep" | "loading";

  // cta abstraction
  cta?: {
    text: string;
    type: "button" | "link";

    // if it is a direct link
    url?: string;
    target?: string;

    // callback when we need to take action like open pylon
    onClick?: () => void;
  };
}
