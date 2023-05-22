package email

import (
	"fmt"
)

type Client struct {
	sender      Sender
	frontendURL string
}

func New(sender Sender, frontendURL string) *Client {
	return &Client{
		sender:      sender,
		frontendURL: frontendURL,
	}
}

func (c *Client) SendOrganizationInvite(toEmail, toName, orgName, roleName string) error {
	err := c.sender.Send(
		toEmail,
		toName,
		"Invitation to join Rill",
		fmt.Sprintf("You have been invited to organization <b>%s</b> as <b>%s</b>. <a href=\"%s\">Please sign into Rill Cloud here to accept invitation</a>.", orgName, roleName, c.frontendURL),
	)
	return err
}

func (c *Client) SendOrganizationAdditionNotification(toEmail, toName, orgName, roleName string) error {
	err := c.sender.Send(
		toEmail,
		toName,
		fmt.Sprintf("You've been added to %q", orgName),
		fmt.Sprintf("You've been added to the organization <b>%s</b> as <b>%s</b>. <a href=\"%s\">This link will take you to your account home in Rill Cloud</a>.", orgName, roleName, c.frontendURL),
	)
	return err
}

func (c *Client) SendProjectInvite(toEmail, toName, projectName, roleName string) error {
	err := c.sender.Send(
		toEmail,
		toName,
		"Invitation to join Rill",
		fmt.Sprintf("You have been invited to project <b>%s</b> as <b>%s</b>. <a href=\"%s\">Please sign into Rill Cloud here to accept invitation</a>.", projectName, roleName, c.frontendURL),
	)
	return err
}

func (c *Client) SendProjectAdditionNotification(toEmail, toName, projectName, roleName string) error {
	err := c.sender.Send(
		toEmail,
		toName,
		fmt.Sprintf("You've been added to %q", projectName),
		fmt.Sprintf("You've been added to the project <b>%s</b> as <b>%s</b>. <a href=\"%s\">This link will take you to your account home in Rill Cloud</a>.", projectName, roleName, c.frontendURL),
	)
	return err
}
