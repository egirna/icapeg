PORTNAME=       	icapeg
DISTVERSIONPREFIX=  v
DISTVERSION=    	1.0.0
CATEGORIES=     	www

MAINTAINER=     	support@egirna.com
COMMENT=        	An ICAP server implementation
WWW=            	https://github.com/egirna/icapeg

LICENSE=        	APACHE20
LICENSE_FILE=   	${WRKSRC}/LICENSE

USES=           	go:modules
GO_MODULE=			github.com/egirna/icapeg

post-install:
	${MKDIR} ${STAGEDIR}${ETCDIR}
	${INSTALL_DATA} ${WRKSRC}/config.toml ${STAGEDIR}${ETCDIR}/config.toml.sample
	${MKDIR} ${STAGEDIR}${DATADIR}
	${INSTALL_DATA} ${WRKSRC}/temp/exception-page.html ${STAGEDIR}${DATADIR}/exception-page.html

PLIST_FILES=    	bin/icapeg \
					"@sample ${ETCDIR}/config.toml.sample" \
					"${DATADIR}/exception-page.html"

.include <bsd.port.mk>
