<?xml version="1.0" encoding="UTF-8"?>
<!--
  Koiné Transit XSLT (SPIKE)
  
  Consumes TransitChart XML. Computes:
  - Transit-to-natal aspects (5 Ptolemaic, whole-sign aware)
  - Transit house overlays (whole-sign)
  - Triplicity dignity of transiting planets
  - Sect-aware interpretation
  
  Does NOT compute:
  - Aspect angles (XSLT does that from longitudes)
  - Orb thresholds (configurable via $orb parameter)
-->
<xsl:stylesheet xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
                xmlns:math="http://exslt.org/math"
                xmlns:exslt="http://exslt.org/common"
                xmlns:koine="urn:empirical:koine"
                version="1.0"
                exclude-result-prefixes="math exslt">

  <!-- ── Parameters ────────────────────────────────────────────────────── -->
  <xsl:param name="orb" select="3.0"/>  <!-- tighter than natal (5°) — transits are time-sensitive -->

  <!-- ── Classical planets ─────────────────────────────────────────────── -->
  <xsl:variable name="classicalPlanets" select="',Sun,Moon,Mercury,Venus,Mars,Jupiter,Saturn,'"/>

  <!-- ── Ptolemaic aspects ────────────────────────────────────────────── -->
  <xsl:variable name="aspects">
    <aspect angle="0" name="conjunction"/>
    <aspect angle="60" name="sextile"/>
    <aspect angle="90" name="square"/>
    <aspect angle="120" name="trine"/>
    <aspect angle="180" name="opposition"/>
  </xsl:variable>

  <!-- ── Sign names ────────────────────────────────────────────────────── -->
  <xsl:variable name="signs" select="'Aries,Taurus,Gemini,Cancer,Leo,Virgo,Libra,Scorpio,Sagittarius,Capricorn,Aquarius,Pisces'"/>

  <!-- ── Domicile rulers ───────────────────────────────────────────────── -->
  <xsl:variable name="domicileSun" select="4"/>
  <xsl:variable name="domicileMoon" select="3"/>
  <xsl:variable name="domicileMercury" select="2"/>
  <xsl:variable name="domicileVenus" select="1"/>
  <xsl:variable name="domicileMars" select="0"/>
  <xsl:variable name="domicileJupiter" select="8"/>
  <xsl:variable name="domicileSaturn" select="9"/>

  <!-- ── Triplicity rulers ─────────────────────────────────────────────── -->
  <xsl:variable name="triplicityFireDay" select="'Sun'"/>
  <xsl:variable name="triplicityFireNight" select="'Jupiter'"/>
  <xsl:variable name="triplicityFirePart" select="'Saturn'"/>
  <xsl:variable name="triplicityEarthDay" select="'Venus'"/>
  <xsl:variable name="triplicityEarthNight" select="'Moon'"/>
  <xsl:variable name="triplicityEarthPart" select="'Mars'"/>
  <xsl:variable name="triplicityAirDay" select="'Saturn'"/>
  <xsl:variable name="triplicityAirNight" select="'Mercury'"/>
  <xsl:variable name="triplicityAirPart" select="'Jupiter'"/>
  <xsl:variable name="triplicityWaterDay" select="'Venus'"/>
  <xsl:variable name="triplicityWaterNight" select="'Mars'"/>
  <xsl:variable name="triplicityWaterPart" select="'Moon'"/>

  <!-- ══════════════════════════════════════════════════════════════════════
       ROOT TEMPLATE
       ════════════════════════════════════════════════════════════════════ -->
  <xsl:template match="/TransitChart">
    <koine:TransitReport xmlns:koine="urn:empirical:koine">
      <koine:Name><xsl:value-of select="Identity/Name"/></koine:Name>
      <koine:TransitDate>
        <xsl:value-of select="Moment/Year"/>-<xsl:value-of select="format-number(Moment/Month,'00')"/>-<xsl:value-of select="format-number(Moment/Day,'00')"/>
      </koine:TransitDate>

      <!-- Determine sect from transit moment -->
      <xsl:variable name="sunLon" select="Transits/Planet[@name='Sun']/Tropical/Lon"/>
      <xsl:variable name="ascLon" select="Angles/ASC"/>
      <xsl:variable name="sunAscDiff" select="($sunLon - $ascLon + 360) mod 360"/>
      <xsl:variable name="isDay" select="$sunAscDiff &lt; 180"/>

      <!-- Capture natal ASC at root level (needed in nested for-each context) -->
      <xsl:variable name="natalASC" select="Natal/Angles/ASC"/>

      <koine:Sect>
        <xsl:choose>
          <xsl:when test="$isDay">day</xsl:when>
          <xsl:otherwise>night</xsl:otherwise>
        </xsl:choose>
      </koine:Sect>

      <!-- ══════════════════════════════════════════════════════════════════
           TRANSIT-TO-NATAL ASPECTS
           ════════════════════════════════════════════════════════════════ -->
      <koine:TransitAspects>
        <xsl:for-each select="Transits/Planet[contains($classicalPlanets, concat(',',@name,','))]">
          <xsl:variable name="tPlanet" select="@name"/>
          <xsl:variable name="tLon" select="Tropical/Lon"/>
          <xsl:variable name="tSign" select="floor($tLon div 30)"/>

          <xsl:for-each select="/TransitChart/Natal/Positions/Planet[contains($classicalPlanets, concat(',',@name,','))]">
            <xsl:variable name="nPlanet" select="@name"/>
            <xsl:variable name="nLon" select="Tropical/Lon"/>
            <xsl:variable name="nSign" select="floor($nLon div 30)"/>

            <!-- Compute angular distance -->
            <xsl:variable name="rawDist" select="$tLon - $nLon"/>
            <xsl:variable name="dist">
              <xsl:choose>
                <xsl:when test="$rawDist &lt; 0"><xsl:value-of select="$rawDist + 360"/></xsl:when>
                <xsl:otherwise><xsl:value-of select="$rawDist"/></xsl:otherwise>
              </xsl:choose>
            </xsl:variable>

            <!-- Check each aspect -->
            <xsl:for-each select="exslt:node-set($aspects)/aspect">
              <xsl:variable name="aspAngle" select="@angle"/>
              <xsl:variable name="aspName" select="@name"/>
              <xsl:variable name="diff">
                <xsl:choose>
                  <xsl:when test="$dist > $aspAngle">
                    <xsl:value-of select="$dist - $aspAngle"/>
                  </xsl:when>
                  <xsl:otherwise>
                    <xsl:value-of select="$aspAngle - $dist"/>
                  </xsl:otherwise>
                </xsl:choose>
              </xsl:variable>

              <xsl:if test="$diff &lt;= $orb">
                <koine:TransitAspect>
                  <koine:TransitPlanet><xsl:value-of select="$tPlanet"/></koine:TransitPlanet>
                  <koine:NatalPlanet><xsl:value-of select="$nPlanet"/></koine:NatalPlanet>
                  <koine:Aspect><xsl:value-of select="$aspName"/></koine:Aspect>
                  <koine:Orb><xsl:value-of select="format-number($diff,'0.00')"/></koine:Orb>
                  <koine:TransitSign>
                    <xsl:call-template name="signName">
                      <xsl:with-param name="idx" select="$tSign"/>
                    </xsl:call-template>
                  </koine:TransitSign>
                  <koine:NatalSign>
                    <xsl:call-template name="signName">
                      <xsl:with-param name="idx" select="$nSign"/>
                    </xsl:call-template>
                  </koine:NatalSign>
                  <!-- Whole-sign house overlay: which natal house is the transit in? -->
                  <koine:NatalHouse>
                    <xsl:call-template name="wholeSignHouse">
                      <xsl:with-param name="lon" select="$tLon"/>
                      <xsl:with-param name="asc" select="$natalASC"/>
                    </xsl:call-template>
                  </koine:NatalHouse>
                </koine:TransitAspect>
              </xsl:if>
            </xsl:for-each>
          </xsl:for-each>
        </xsl:for-each>
      </koine:TransitAspects>

      <!-- ══════════════════════════════════════════════════════════════════
           TRANSIT PLANET DIGNITY (triplicity)
           ════════════════════════════════════════════════════════════════ -->
      <koine:TransitDignities>
        <xsl:for-each select="Transits/Planet[contains($classicalPlanets, concat(',',@name,','))]">
          <xsl:variable name="planet" select="@name"/>
          <xsl:variable name="lon" select="Tropical/Lon"/>
          <xsl:variable name="sign" select="floor($lon div 30)"/>

          <koine:TransitDignity>
            <koine:Planet><xsl:value-of select="$planet"/></koine:Planet>
            <koine:Sign>
              <xsl:call-template name="signName">
                <xsl:with-param name="idx" select="$sign"/>
              </xsl:call-template>
            </koine:Sign>
            <koine:TriplicityRuler>
              <xsl:call-template name="triplicityRuler">
                <xsl:with-param name="sign" select="$sign"/>
                <xsl:with-param name="isDay" select="$isDay"/>
              </xsl:call-template>
            </koine:TriplicityRuler>
            <koine:IsTriplicityRuler>
              <xsl:call-template name="isTriplicityRuler">
                <xsl:with-param name="planet" select="$planet"/>
                <xsl:with-param name="sign" select="$sign"/>
                <xsl:with-param name="isDay" select="$isDay"/>
              </xsl:call-template>
            </koine:IsTriplicityRuler>
          </koine:TransitDignity>
        </xsl:for-each>
      </koine:TransitDignities>

    </koine:TransitReport>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       HELPER TEMPLATES
       ════════════════════════════════════════════════════════════════════ -->

  <xsl:template name="signName">
    <xsl:param name="idx"/>
    <xsl:call-template name="nthToken">
      <xsl:with-param name="list" select="$signs"/>
      <xsl:with-param name="n" select="$idx + 1"/>
    </xsl:call-template>
  </xsl:template>

  <!-- Whole-sign house: house 1 = sign containing ASC, house 2 = next sign, etc. -->
  <xsl:template name="wholeSignHouse">
    <xsl:param name="lon"/>
    <xsl:param name="asc"/>
    <xsl:variable name="sign" select="floor($lon div 30)"/>
    <xsl:variable name="ascSign" select="floor($asc div 30)"/>
    <xsl:variable name="house" select="(($sign - $ascSign + 12) mod 12) + 1"/>
    <xsl:value-of select="$house"/>
  </xsl:template>

  <xsl:template name="triplicityRuler">
    <xsl:param name="sign"/>
    <xsl:param name="isDay"/>
    <xsl:variable name="element" select="$sign mod 4"/>
    <!-- 0=Fire, 1=Earth, 2=Air, 3=Water -->
    <xsl:choose>
      <xsl:when test="$element = 0">
        <xsl:choose>
          <xsl:when test="$isDay"><xsl:value-of select="$triplicityFireDay"/></xsl:when>
          <xsl:otherwise><xsl:value-of select="$triplicityFireNight"/></xsl:otherwise>
        </xsl:choose>
      </xsl:when>
      <xsl:when test="$element = 1">
        <xsl:choose>
          <xsl:when test="$isDay"><xsl:value-of select="$triplicityEarthDay"/></xsl:when>
          <xsl:otherwise><xsl:value-of select="$triplicityEarthNight"/></xsl:otherwise>
        </xsl:choose>
      </xsl:when>
      <xsl:when test="$element = 2">
        <xsl:choose>
          <xsl:when test="$isDay"><xsl:value-of select="$triplicityAirDay"/></xsl:when>
          <xsl:otherwise><xsl:value-of select="$triplicityAirNight"/></xsl:otherwise>
        </xsl:choose>
      </xsl:when>
      <xsl:otherwise>
        <xsl:choose>
          <xsl:when test="$isDay"><xsl:value-of select="$triplicityWaterDay"/></xsl:when>
          <xsl:otherwise><xsl:value-of select="$triplicityWaterNight"/></xsl:otherwise>
        </xsl:choose>
      </xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <xsl:template name="isTriplicityRuler">
    <xsl:param name="planet"/>
    <xsl:param name="sign"/>
    <xsl:param name="isDay"/>
    <xsl:variable name="ruler">
      <xsl:call-template name="triplicityRuler">
        <xsl:with-param name="sign" select="$sign"/>
        <xsl:with-param name="isDay" select="$isDay"/>
      </xsl:call-template>
    </xsl:variable>
    <xsl:choose>
      <xsl:when test="$planet = $ruler">true</xsl:when>
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- nthToken: extract the n-th comma-delimited token (1-indexed) -->
  <xsl:template name="nthToken">
    <xsl:param name="list"/>
    <xsl:param name="n"/>
    <xsl:param name="delimiter" select="','"/>
    <xsl:choose>
      <xsl:when test="$n = 1">
        <xsl:choose>
          <xsl:when test="contains($list, $delimiter)">
            <xsl:value-of select="substring-before($list, $delimiter)"/>
          </xsl:when>
          <xsl:otherwise>
            <xsl:value-of select="$list"/>
          </xsl:otherwise>
        </xsl:choose>
      </xsl:when>
      <xsl:when test="contains($list, $delimiter)">
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="substring-after($list, $delimiter)"/>
          <xsl:with-param name="n" select="$n - 1"/>
          <xsl:with-param name="delimiter" select="$delimiter"/>
        </xsl:call-template>
      </xsl:when>
    </xsl:choose>
  </xsl:template>

</xsl:stylesheet>
