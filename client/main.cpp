#include "platform_proxy.h"

#include <QApplication>
#include <QByteArray>
#include <QCloseEvent>
#include <QCoreApplication>
#include <QDateTime>
#include <QFormLayout>
#include <QGroupBox>
#include <QHBoxLayout>
#include <QLabel>
#include <QLineEdit>
#include <QMainWindow>
#include <QMessageBox>
#include <QPlainTextEdit>
#include <QProcess>
#include <QPushButton>
#include <QRegularExpression>
#include <QRegularExpressionValidator>
#include <QSpinBox>
#include <QTcpSocket>
#include <QTextBrowser>
#include <QTimer>
#include <QVBoxLayout>
#include <QWidget>

class SnekProxyWindow final : public QMainWindow {
public:
    SnekProxyWindow() {
        setWindowTitle("SnekProxy Client");
        resize(760, 700);

        auto* central = new QWidget(this);
        auto* layout = new QVBoxLayout(central);
        auto* intro = new QLabel(
            "<h2>Secure proxy connection</h2>"
            "<p>Create an encrypted SSH tunnel to a host. SnekProxy will configure the current "
            "user's system proxy while the tunnel is active.</p>",
            central);
        intro->setWordWrap(true);
        layout->addWidget(intro);

        auto* connection_box = new QGroupBox("Connection details", central);
        auto* form = new QFormLayout(connection_box);
        host_ = new QLineEdit(connection_box);
        host_->setPlaceholderText("Example: 192.168.1.20 or proxy.example.com");
        host_->setValidator(new QRegularExpressionValidator(
            QRegularExpression("^(?:[A-Za-z0-9](?:[A-Za-z0-9.-]*[A-Za-z0-9])?|[0-9A-Fa-f:]+)$"),
            host_));
        user_ = new QLineEdit(connection_box);
        user_->setPlaceholderText("Regular account on the proxy host");
        user_->setValidator(new QRegularExpressionValidator(
            QRegularExpression("^[A-Za-z_][A-Za-z0-9_.-]*$"), user_));
        ssh_port_ = new QSpinBox(connection_box);
        ssh_port_->setRange(1, 65535);
        ssh_port_->setValue(22);
        local_port_ = new QSpinBox(connection_box);
        local_port_->setRange(1, 65535);
        local_port_->setValue(1080);
        proxy_address_ = new QLabel("127.0.0.1:1080", connection_box);
        form->addRow("Host IP or DNS name:", host_);
        form->addRow("SSH username:", user_);
        form->addRow("SSH port:", ssh_port_);
        form->addRow("Local SOCKS5 port:", local_port_);
        form->addRow("Proxy address for your apps:", proxy_address_);
        connect(local_port_, qOverload<int>(&QSpinBox::valueChanged), this,
                [this](int port) { proxy_address_->setText(QString("127.0.0.1:%1").arg(port)); });
        layout->addWidget(connection_box);

        auto* auth_note = new QLabel(
            "Authentication uses your existing SSH keys/ssh-agent. For safety, the client does "
            "not collect or save passwords and does not automatically trust new host keys.",
            central);
        auth_note->setWordWrap(true);
        layout->addWidget(auth_note);

        auto* buttons = new QHBoxLayout();
        connect_button_ = new QPushButton("Connect", central);
        disable_button_ = new QPushButton("Disable Tunnel", central);
        disable_button_->setEnabled(false);
        buttons->addWidget(connect_button_);
        buttons->addWidget(disable_button_);
        buttons->addStretch();
        layout->addLayout(buttons);

        status_ = new QLabel("Not connected", central);
        layout->addWidget(status_);

        auto* steps = new QTextBrowser(central);
        steps->setMaximumHeight(175);
        steps->setHtml(
            "<h3>What happens when you connect</h3>"
            "<ol>"
            "<li>Enter the host address and a regular SSH account (not root).</li>"
            "<li>The host key must already be verified in your SSH <code>known_hosts</code> file. "
            "First connect from a terminal with <code>ssh -p PORT USER@HOST</code> and verify its "
            "fingerprint with the host administrator.</li>"
            "<li>SnekProxy starts encrypted SSH SOCKS5 forwarding, then configures your current "
            "user's system proxy automatically.</li>"
            "<li>Click <b>Disable Tunnel</b> to stop SSH and restore the proxy settings that were "
            "active before connecting.</li>"
            "</ol>"
            "<p>GNOME Linux and Windows system proxy settings are supported. Some applications "
            "may ignore system proxy settings. This is not a system-wide kill switch.</p>");
        layout->addWidget(steps);

        output_ = new QPlainTextEdit(central);
        output_->setReadOnly(true);
        output_->setPlaceholderText("SSH and proxy setup messages will appear here.");
        output_->setMaximumBlockCount(500);
        layout->addWidget(output_, 1);
        setCentralWidget(central);

        process_ = new QProcess(this);
        proxy_settings_ = new ProxySettings();
        probe_timer_ = new QTimer(this);
        probe_timer_->setInterval(250);
        connect_button_->setDefault(true);

        connect(connect_button_, &QPushButton::clicked, this, [this] { startTunnel(); });
        connect(disable_button_, &QPushButton::clicked, this, [this] { disableTunnel(); });
        connect(probe_timer_, &QTimer::timeout, this, [this] { probeLocalSocks(); });
        connect(process_, &QProcess::started, this, [this] {
            connect_button_->setEnabled(false);
            disable_button_->setEnabled(true);
            proxy_deadline_ = QDateTime::currentMSecsSinceEpoch() + 10000;
            status_->setText("SSH started. Waiting for the local SOCKS5 tunnel…");
            probe_timer_->start();
        });
        connect(process_, &QProcess::readyReadStandardOutput, this, [this] {
            appendOutput(process_->readAllStandardOutput());
        });
        connect(process_, &QProcess::readyReadStandardError, this, [this] {
            appendOutput(process_->readAllStandardError());
        });
        connect(process_, &QProcess::errorOccurred, this, [this](QProcess::ProcessError error) {
            if (error == QProcess::FailedToStart) {
                probe_timer_->stop();
                status_->setText("Could not start SSH. Install the OpenSSH client and try again.");
                connect_button_->setEnabled(true);
                disable_button_->setEnabled(proxy_settings_->hasPendingRestore());
                QMessageBox::critical(this, "SSH client unavailable",
                                       "Could not start 'ssh'. Install the OpenSSH client.");
            }
        });
        connect(process_, qOverload<int, QProcess::ExitStatus>(&QProcess::finished),
                this, [this](int exit_code, QProcess::ExitStatus exit_status) {
                    probe_timer_->stop();
                    QString restore_error;
                    const bool restored = proxy_settings_->restore(restore_error);
                    connect_button_->setEnabled(true);
                    disable_button_->setEnabled(!restored);
                    if (!restored) {
                        status_->setText("Tunnel ended, but proxy settings still need restoration.");
                        output_->appendPlainText("Proxy restoration error: " + restore_error);
                    } else {
                        status_->setText(exit_status == QProcess::NormalExit && exit_code == 0
                                             ? "Tunnel disabled; previous proxy settings restored."
                                             : QString("SSH exited (code %1); previous proxy settings restored.")
                                                   .arg(exit_code));
                    }
                });

        if (proxy_settings_->hasPendingRestore()) {
            QString error;
            if (proxy_settings_->restore(error)) {
                status_->setText("Recovered and restored settings left by an earlier run.");
            } else {
                status_->setText("Proxy settings require restoration. Use Disable Tunnel.");
                disable_button_->setEnabled(true);
                output_->appendPlainText("Recovery error: " + error);
            }
        }
    }

private:
    void closeEvent(QCloseEvent* event) override {
        if (process_->state() != QProcess::NotRunning || proxy_settings_->hasPendingRestore()) {
            const auto answer = QMessageBox::question(
                this, "Disable proxy tunnel?",
                "Closing SnekProxy will stop the SSH tunnel and restore your previous system "
                "proxy settings. Continue?",
                QMessageBox::Yes | QMessageBox::Cancel, QMessageBox::Yes);
            if (answer != QMessageBox::Yes) {
                event->ignore();
                return;
            }

            QString error;
            if (!proxy_settings_->restore(error)) {
                QMessageBox::critical(this, "Could not restore proxy settings",
                                      "The SSH tunnel will be stopped, but proxy settings could "
                                      "not be restored:\n\n" + error);
            }
            if (process_->state() != QProcess::NotRunning) {
                process_->terminate();
                if (!process_->waitForFinished(3000)) {
                    process_->kill();
                    process_->waitForFinished();
                }
            }
        }
        event->accept();
    }

    void startTunnel() {
        const auto host = host_->text().trimmed();
        const auto username = user_->text().trimmed();
        if (host.isEmpty() || !host_->hasAcceptableInput()) {
            QMessageBox::warning(this, "Check host", "Enter a valid host IP address or DNS name.");
            host_->setFocus();
            return;
        }
        if (username.isEmpty() || !user_->hasAcceptableInput()) {
            QMessageBox::warning(this, "Check username", "Enter a valid regular SSH username.");
            user_->setFocus();
            return;
        }
        if (proxy_settings_->hasPendingRestore()) {
            QString error;
            if (!proxy_settings_->restore(error)) {
                QMessageBox::critical(this, "Proxy settings need recovery",
                                      "Restore the previous system proxy settings before starting "
                                      "a new tunnel:\n\n" + error);
                return;
            }
        }

        const auto destination = username + "@" + host;
        const QStringList arguments{
            "-N",
            "-D", QString("127.0.0.1:%1").arg(local_port_->value()),
            "-p", QString::number(ssh_port_->value()),
            "-o", "ExitOnForwardFailure=yes",
            "-o", "ServerAliveInterval=30",
            "-o", "ServerAliveCountMax=3",
            "-o", "BatchMode=yes",
            "-o", "StrictHostKeyChecking=yes",
            "--", destination,
        };
        output_->clear();
        output_->appendPlainText("Starting SSH tunnel to " + destination);
        output_->appendPlainText(
            QString("Local SOCKS5 address: 127.0.0.1:%1").arg(local_port_->value()));
        status_->setText("Starting SSH…");
        connect_button_->setEnabled(false);
        process_->setProgram("ssh");
        process_->setArguments(arguments);
        process_->start();
    }

    void probeLocalSocks() {
        if (process_->state() == QProcess::NotRunning) {
            probe_timer_->stop();
            return;
        }
        QTcpSocket socket;
        socket.connectToHost("127.0.0.1", static_cast<quint16>(local_port_->value()));
        if (socket.waitForConnected(100)) {
            socket.write(QByteArray::fromHex("050100"));
            bool greeted = socket.waitForBytesWritten(100);
            QByteArray response;
            while (greeted && response.size() < 2) {
                if (!socket.waitForReadyRead(100)) {
                    greeted = false;
                    break;
                }
                response.append(socket.read(2 - response.size()));
            }
            greeted = greeted && response == QByteArray::fromHex("0500");
            socket.abort();
            if (!greeted) {
                if (QDateTime::currentMSecsSinceEpoch() >= proxy_deadline_) {
                    probe_timer_->stop();
                    status_->setText("Local port is not responding as an SSH SOCKS5 proxy.");
                    output_->appendPlainText(
                        "The local SOCKS5 check failed; system proxy was not changed.");
                    process_->terminate();
                }
                return;
            }
            probe_timer_->stop();
            QString error;
            if (!proxy_settings_->apply(local_port_->value(), error)) {
                status_->setText("Could not set the system proxy; stopping tunnel.");
                output_->appendPlainText("Proxy setup error: " + error);
                process_->terminate();
                QTimer::singleShot(3000, process_, [this] {
                    if (process_->state() != QProcess::NotRunning) {
                        process_->kill();
                    }
                });
                return;
            }
            status_->setText("Connected. Current-user system proxy is enabled.");
            output_->appendPlainText(
                "System proxy settings updated. Use Disable Tunnel to restore them.");
            disable_button_->setEnabled(true);
            return;
        }
        if (QDateTime::currentMSecsSinceEpoch() >= proxy_deadline_) {
            probe_timer_->stop();
            status_->setText("SSH did not establish the local SOCKS5 tunnel.");
            output_->appendPlainText(
                "No SOCKS listener appeared within 10 seconds; system proxy was not changed.");
            process_->terminate();
        }
    }

    void disableTunnel() {
        probe_timer_->stop();
        QString error;
        const bool restored = proxy_settings_->restore(error);
        if (!restored) {
            output_->appendPlainText("Proxy restoration error: " + error);
            status_->setText("Proxy settings still need restoration.");
        } else {
            status_->setText("Previous system proxy settings restored.");
        }
        if (process_->state() != QProcess::NotRunning) {
            process_->terminate();
            QTimer::singleShot(3000, process_, [this] {
                if (process_->state() != QProcess::NotRunning) {
                    process_->kill();
                }
            });
        } else {
            connect_button_->setEnabled(true);
            disable_button_->setEnabled(!restored);
        }
    }

    void appendOutput(const QByteArray& bytes) {
        const auto text = QString::fromLocal8Bit(bytes).trimmed();
        if (!text.isEmpty()) {
            output_->appendPlainText(text);
        }
    }

    QLineEdit* host_{};
    QLineEdit* user_{};
    QSpinBox* ssh_port_{};
    QSpinBox* local_port_{};
    QLabel* proxy_address_{};
    QPushButton* connect_button_{};
    QPushButton* disable_button_{};
    QLabel* status_{};
    QPlainTextEdit* output_{};
    QProcess* process_{};
    QTimer* probe_timer_{};
    ProxySettings* proxy_settings_{};
    qint64 proxy_deadline_{};
};

int main(int argc, char* argv[]) {
    QApplication application(argc, argv);
    QCoreApplication::setOrganizationName("SnekProxy");
    QCoreApplication::setApplicationName("SnekProxy Client");
    SnekProxyWindow window;
    window.show();
    return application.exec();
}
