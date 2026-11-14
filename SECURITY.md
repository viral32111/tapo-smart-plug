# Security Policy

The security of my projects and the people who use them is important to me. I appreciate the time and effort taken by security researchers and members of the community to identify and responsibly report potential vulnerabilities.

As an independent developer maintaining these projects in my own time, my ability to investigate reports, develop fixes, and provide ongoing security support is limited. I will endeavour to handle legitimate security reports responsibly and address issues as soon as I reasonably can, but I cannot guarantee response times, resolution times, or ongoing maintenance.

This policy outlines how to report security vulnerabilities and what to expect from the disclosure and remediation process.

## Reporting a Vulnerability

**Please report suspected security vulnerabilities privately. Do not disclose vulnerabilities through public GitHub issues, discussions, pull requests, or other public channels.**

Security reports should be submitted using the contact details provided on my website:

- **Contact details:** [viral32111.com/contact](https://viral32111.com/contact)
- **PGP public key:** [viral32111.com/public.txt](https://viral32111.com/public.txt)

For sensitive reports, I strongly recommend encrypting your message using my PGP public key.

If you choose to encrypt your report, please include your own PGP public key, or provide a way for me to obtain it, so that I can encrypt any replies.

Please do not submit sensitive vulnerability details through public communication channels.

### What to Include

To help me understand and investigate a report, please include as much of the following information as possible:

- **Project:** The name and repository of the affected project.
- **Affected version:** The version, release, commit, or revision where the vulnerability was discovered.
- **Description:** A clear explanation of the vulnerability and its potential security implications.
- **Reproduction steps:** Instructions for reproducing the issue, preferably with a minimal proof of concept.
- **Impact:** What an attacker could potentially achieve by exploiting the vulnerability, including any affected users, systems, or data.
- **Environment:** Any relevant operating systems, software versions, configurations, dependencies, or other environmental details.
- **Suggested remediation:** Any recommendations you may have for addressing the vulnerability.

Reports do not need to follow a particular format. However, providing sufficient information to reproduce and understand the issue will make investigation considerably easier.

## Responsible Disclosure

I ask that security researchers follow responsible disclosure practices when investigating and reporting vulnerabilities.

In particular:

- Allow me a reasonable opportunity to acknowledge, investigate, and address a reported vulnerability before publicly disclosing it.
- Do not publicly share vulnerability details, proof-of-concept exploits, or other information that could enable exploitation before a fix or an appropriate disclosure arrangement has been reached.
- Avoid accessing, modifying, deleting, or disclosing data that does not belong to you.
- Do not intentionally disrupt services, degrade availability, or interfere with the operation of systems or infrastructure.
- Do not perform testing against systems, services, or infrastructure that you do not own or have explicit permission to test.
- Do not use vulnerabilities to compromise other users, escalate access beyond what is necessary to demonstrate an issue, or otherwise cause harm.

Please keep testing proportionate and limited to what is reasonably necessary to demonstrate the reported vulnerability.

If you accidentally encounter sensitive information or gain unintended access to data, stop testing and include the relevant details in your private report.

I appreciate researchers who make a genuine effort to protect users and work collaboratively towards resolving security issues.

## Response and Remediation

I will make a reasonable effort to acknowledge legitimate vulnerability reports and investigate them as my availability permits.

However, **I am an individual developer, not a dedicated security team.** I maintain my projects alongside other personal and professional commitments, and I cannot guarantee that I will be able to respond to every report immediately.

Please bear the following in mind:

- There are no guaranteed response, investigation, or remediation deadlines.
- The time required to investigate and resolve an issue will depend on its severity, complexity, and the availability of the information needed to reproduce it.
- Some reports may require considerable research, development, testing, or changes to project architecture before a suitable fix can be released.
- I may need to request additional information or clarification before proceeding.
- In some circumstances, I may be unable to reproduce, investigate, or resolve a reported issue.

I will endeavour to communicate meaningful progress where possible, but please understand that periods of limited communication do not necessarily mean a report has been ignored.

### Fixes and Supported Versions

When a vulnerability is confirmed and a suitable fix is available, I will endeavour to release an updated version of the affected project.

My preferred approach is to address security issues in a new release rather than maintaining and patching every previously published version.

Consequently:

- Security fixes will most commonly be included in a new version or release.
- Previous versions may not receive individual security updates or backported fixes.
- Whether older versions receive a fix will depend on the project, the severity of the vulnerability, and the effort involved.
- I cannot guarantee that every affected version will be updated.
- Users may need to upgrade to a newer release to benefit from a security fix.

Where practical, relevant information about a vulnerability will be disclosed after a fix has been made available. This may include a description of the issue, affected versions, the fixed version, and any recommended actions for users.

I may also use GitHub Security Advisories or other appropriate channels to communicate security-related information where available and practical.

## Vulnerability Severity and Prioritisation

I will consider the potential impact and severity of each report when deciding how to prioritise investigation and remediation.

Factors may include:

- The ease with which the vulnerability can be exploited.
- Whether exploitation requires authentication or specific privileges.
- The potential impact on the confidentiality, integrity, or availability of data and systems.
- The number of users, installations, or services potentially affected.
- Whether the vulnerability is actively being exploited or presents an immediate risk.

I may use established severity classification systems, such as the [Common Vulnerability Scoring System (CVSS)](https://www.first.org/cvss/), as a reference when assessing reported vulnerabilities.

However, severity classifications and suggested priorities do not guarantee a particular response or remediation timeframe.

## Rewards and Recognition

I do not operate a formal bug bounty programme and cannot offer monetary rewards for vulnerability reports.

I maintain these projects independently and do not have the resources to provide financial compensation for security research.

Nevertheless, I genuinely appreciate the effort involved in identifying and responsibly reporting security issues.

Depending on the project, the nature of the vulnerability, and the circumstances of the report, I may be able to offer a non-monetary reward.

At the very least, I will endeavour to provide appropriate recognition, credit, or public thanks to researchers who have helped improve the security of my projects.

Any public acknowledgement will be subject to the researcher's wishes. Please let me know if you would prefer to remain anonymous or do not want your contribution publicly recognised.

There is no guarantee that a reward or public acknowledgement will be provided in every case.

## Scope and Eligibility

This policy applies to the projects I personally maintain and publish, including their associated source code, releases, packages, applications, and other components where applicable.

Due to the variety of projects I work on, there is no single universal definition of what is considered in scope. Please use reasonable judgement when determining whether an issue is relevant to a particular project.

In general, security vulnerabilities that have a demonstrable impact on the security, integrity, or availability of a project or its users are welcome.

The following are generally not considered security vulnerabilities eligible for this disclosure process on their own:

- Issues that do not have a demonstrable security impact.
- Theoretical vulnerabilities without a practical or credible exploitation scenario.
- Reports concerning outdated software or dependencies where no exploitable vulnerability in the project has been demonstrated.
- Security best-practice recommendations without an identifiable vulnerability.
- Issues that require unrealistic assumptions or highly unlikely circumstances to exploit.

This is not an exhaustive list. Individual reports will be considered on their own merits.

Third-party services, dependencies, platforms, and infrastructure may be outside my control. If a vulnerability originates in a third-party product or service, I may be unable to resolve it directly and may instead refer the report to the appropriate maintainer.

## Policy Limitations

This policy describes my intentions and the approach I will endeavour to follow when handling security reports.

It should not be interpreted as a guarantee of ongoing security maintenance, support for any particular project or version, a guaranteed response or resolution timeframe, or an obligation to provide rewards or compensation.

Projects may be discontinued, archived, abandoned, or otherwise become unmaintained. In such circumstances, my ability to investigate or address reported vulnerabilities may be significantly limited or unavailable.

I appreciate your understanding of these limitations and your consideration when deciding whether to rely on any of my projects in security-sensitive environments.

## Contact

For all security-related enquiries and vulnerability reports, please use the contact details provided on my website: [viral32111.com/contact](https://viral32111.com/contact)

For encrypted communication, please use my [PGP public key](https://viral32111.com/public.txt).

**Please keep vulnerability reports private until an appropriate disclosure arrangement has been reached.**

Thank you for helping make my projects and the wider community a little safer.
