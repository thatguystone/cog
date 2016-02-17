package kafkalocal;

import java.io.FileNotFoundException;
import java.io.IOException;
import java.util.Properties;
import kafka.server.KafkaConfig;
import kafka.server.KafkaServerStartable;
import org.apache.zookeeper.server.quorum.QuorumPeerConfig;
import org.apache.zookeeper.server.ServerConfig;
import org.apache.zookeeper.server.ZooKeeperServerMain;

// Adapted from: https://gist.github.com/fjavieralba/7930018

public class KafkaLocal {
	public static void main(String []args) throws Exception {
		if (args.length != 1) {
			System.out.println("Usage: KafkaLocal <tmpdir>\n");
			System.exit(2);
		}

		startZk(args[0]);
		startKafka(args[0]);
	}

	static Properties loadProps(final String path) throws IOException {
		Properties props = new Properties();
		props.load(Class.class.getResourceAsStream(path));
		return props;
	}

	static void startZk(final String tmpDir) throws Exception {
		Properties props = loadProps("/zk.properties");
		props.setProperty("dataDir", tmpDir + "/zk");

		QuorumPeerConfig quorumCfg = new QuorumPeerConfig();
		quorumCfg.parseProperties(props);

		final ServerConfig cfg = new ServerConfig();
		cfg.readFrom(quorumCfg);

		new Thread() {
			public void run() {
				try {
					new ZooKeeperServerMain().runFromConfig(cfg);
				} catch (Exception e) {
					System.out.println("ZooKeeper exception: " + e);
					System.exit(1);
				}
			}
		}.start();
	}

	static void startKafka(final String tmpDir) throws IOException {
		Properties props = loadProps("/kafka.properties");
		props.setProperty("log.dirs", tmpDir + "/kafka");

		KafkaConfig kafkaConfig = new KafkaConfig(props);

		KafkaServerStartable kafka = new KafkaServerStartable(kafkaConfig);
		kafka.startup();
	}
}
