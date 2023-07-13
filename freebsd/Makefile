PORTNAME=       	icapeg
DISTVERSIONPREFIX=  v
DISTVERSION=    	1.0.0
PORTREVISION=		1
CATEGORIES=     	www

MAINTAINER=     	support@egirna.com
COMMENT=        	ICAP server implementation with Multi vendor support.
WWW=            	https://github.com/egirna/icapeg

LICENSE=        	APACHE20
LICENSE_FILE=   	${WRKSRC}/LICENSE

USES=           	go:modules
GO_MODULE=			github.com/egirna/icapeg

USE_RC_SUBR=		icapeg

post-install:
	${MKDIR} ${STAGEDIR}${ETCDIR}
	${INSTALL_DATA} ${WRKSRC}/config.toml ${STAGEDIR}${ETCDIR}/config.toml.sample
	${MKDIR} ${STAGEDIR}${DATADIR}
	${INSTALL_DATA} ${WRKSRC}/temp/exception-page.html ${STAGEDIR}${DATADIR}/exception-page.html

PLIST_FILES=    	bin/icapeg \
					"@sample ${ETCDIR}/config.toml.sample" \
					"${DATADIR}/exception-page.html"

.include <bsd.port.mk>
